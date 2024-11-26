package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	legacyKey  = "enterprise.cloud.ionos.com/datacenter-id"
	newKey     = "cloud.ionos.com/datacenter-id"
	driverName = "cloud.ionos.com"
	annKey     = "cloud.ionos.com/topology-migration"
)

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")
	var backupDir string
	var dryRun bool

	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "kubeconfig file to use")
	flag.StringVar(&backupDir, "backup-dir", backupDir, "directory to store unmodified persistentvolumes in (defaults to generated tempdir)")
	flag.BoolVar(&dryRun, "dry-run", dryRun, "don't modify objects")
	flag.Parse()

	if err := run(kubeconfig, backupDir, dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(kubeconfig, backupDir string, dryRun bool) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pvs, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list persistentvolumes: %w", err)
	}

	var dryRunOpt []string
	if dryRun {
		dryRunOpt = []string{metav1.DryRunAll}
	}

	if backupDir == "" {
		backupDir, err = os.MkdirTemp("", "migrate-pv-topology-*")
	} else {
		backupDir, err = filepath.Abs(filepath.Clean(backupDir))
	}
	if err != nil {
		return err
	}

	fmt.Printf("writing persistentvolume backups to %s\n", backupDir)

	g, ctx := errgroup.WithContext(ctx)

	for _, pv := range pvs.Items {
		if !shouldProcessPV(pv) {
			continue
		}

		updated := updatePV(pv)
		if updated == nil {
			continue
		}

		var claim *corev1.PersistentVolumeClaim

		if pv.Spec.ClaimRef != nil {
			pvc, err := client.CoreV1().PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(ctx, pv.Spec.ClaimRef.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get persistentvolumeclaim %s/%s: %w", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name, err)
			}

			if pvc.Status.Phase != corev1.ClaimBound {
				fmt.Printf("skipped persistentvolume %s with persistentvolumeclaim %s/%s in phase %s\n", pv.Name, pvc.Namespace, pvc.Name, pvc.Status.Phase)
				continue
			}

			claim = pvc
		}

		if err := writePV(backupDir, pv); err != nil {
			return fmt.Errorf("store persistentvolume backup: %w", err)
		}

		if pv.Spec.PersistentVolumeReclaimPolicy != corev1.PersistentVolumeReclaimRetain {
			patchData := []byte(`{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}`)
			_, err := client.CoreV1().PersistentVolumes().Patch(ctx, pv.Name, types.MergePatchType, patchData, metav1.PatchOptions{DryRun: dryRunOpt})
			if err != nil {
				return fmt.Errorf("patch persistentvolume %s: %w", pv.Name, err)
			}
		}

		err := client.CoreV1().PersistentVolumes().Delete(ctx, pv.Name, metav1.DeleteOptions{DryRun: dryRunOpt})
		if err != nil {
			return fmt.Errorf("delete persistentvolume %s: %w", pv.Name, err)
		}

		patchData := []byte(`{"metadata":{"finalizers":[]}}`)
		_, err = client.CoreV1().PersistentVolumes().Patch(ctx, pv.Name, types.MergePatchType, patchData, metav1.PatchOptions{DryRun: dryRunOpt})
		if err != nil {
			return fmt.Errorf("patch persistentvolume %s: %w", pv.Name, err)
		}

		_, err = client.CoreV1().PersistentVolumes().Create(ctx, updated, metav1.CreateOptions{DryRun: dryRunOpt})
		if err != nil && (!dryRun || !apierrors.IsAlreadyExists(err)) {
			return fmt.Errorf("re-create persistentvolume %s: %w", pv.Name, err)
		}

		if dryRun {
			fmt.Printf("migrated persistentvolume %s (dry-run)\n", pv.Name)
		} else {
			fmt.Printf("migrated persistentvolume %s\n", pv.Name)
		}

		if claim != nil {
			g.Go(func() error {
				return waitForClaimBound(ctx, client, claim)
			})
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func writePV(rootDir string, pv corev1.PersistentVolume) error {
	filename := filepath.Join(rootDir, fmt.Sprintf("%s.json", pv.Name))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// update some server-side fields for easier usage
	pv.CreationTimestamp = metav1.Time{}
	pv.ManagedFields = nil
	pv.ResourceVersion = ""
	pv.UID = ""
	pv.Status = corev1.PersistentVolumeStatus{}
	pv.APIVersion = "v1"
	pv.Kind = "PersistentVolume"

	return json.NewEncoder(f).Encode(pv)
}

func shouldProcessPV(pv corev1.PersistentVolume) bool {
	return pv.Spec.CSI != nil && pv.Spec.CSI.Driver == driverName &&
		pv.Status.Phase != corev1.VolumeAvailable && pv.Status.Phase != corev1.VolumeReleased
}

func updatePV(pv corev1.PersistentVolume) *corev1.PersistentVolume {
	for termIndex, term := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
		for exprIndex, expr := range term.MatchExpressions {
			if expr.Key == legacyKey {
				updated := pv.DeepCopy()
				updated.Spec.NodeAffinity.Required.NodeSelectorTerms[termIndex].MatchExpressions[exprIndex].Key = newKey
				return updated
			}
		}
	}
	return nil
}

func waitForClaimBound(ctx context.Context, client kubernetes.Interface, claim *corev1.PersistentVolumeClaim) error {
	claim, err := client.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(ctx, claim.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get persistentvolumeclaim %s/%s: %w", claim.Namespace, claim.Name, err)
	}

	if claim.Status.Phase == corev1.ClaimBound {
		return nil
	}

	w, err := client.CoreV1().PersistentVolumeClaims(claim.Namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("init watch for persistentvolumeclaim %s/%s: %w", claim.Namespace, claim.Name, err)
	}
	defer w.Stop()

	if _, ok := claim.Annotations[annKey]; !ok {
		// annotate persistentvolumeclaim to try and speed up reconciliation
		if claim.Annotations == nil {
			claim.Annotations = make(map[string]string)
		}
		claim.Annotations[annKey] = ""
		claim, err = client.CoreV1().PersistentVolumeClaims(claim.Namespace).Update(ctx, claim, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("annotate persistentvolumeclaim %s/%s: %w", claim.Namespace, claim.Name, err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w while waiting for persistentvolumeclaim %s/%s to be bound", ctx.Err(), claim.Namespace, claim.Name)
		case ev, closed := <-w.ResultChan():
			if closed {
				return nil
			}

			pvc := ev.Object.(*corev1.PersistentVolumeClaim)
			if pvc.Namespace == claim.Namespace && pvc.Name == claim.Name && pvc.Status.Phase == corev1.ClaimBound {
				claim, err := client.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(ctx, claim.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("get persistentvolumeclaim %s/%s: %w", claim.Namespace, claim.Name, err)
				}

				delete(claim.Annotations, annKey)
				_, err = client.CoreV1().PersistentVolumeClaims(claim.Namespace).Update(ctx, claim, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("annotate persistentvolumeclaim %s/%s: %w", claim.Namespace, claim.Name, err)
				}

				return nil
			}
		}
	}
}
