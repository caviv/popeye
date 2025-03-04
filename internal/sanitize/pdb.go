package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	"github.com/rs/zerolog/log"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	// PodDisruptionBudget tracks PodDisruptionBudget sanitization.
	PodDisruptionBudget struct {
		*issues.Collector
		PodDisruptionBudgetLister
	}

	// PodDisruptionBudgetLister list available PodDisruptionBudgets on a cluster.
	PodDisruptionBudgetLister interface {
		PodLister
		ListPodDisruptionBudgets() map[string]*pv1beta1.PodDisruptionBudget
	}
)

// NewPodDisruptionBudget returns a new PodDisruptionBudget sanitizer.
func NewPodDisruptionBudget(c *issues.Collector, lister PodDisruptionBudgetLister) *PodDisruptionBudget {
	return &PodDisruptionBudget{
		Collector:                 c,
		PodDisruptionBudgetLister: lister,
	}
}

// Sanitize cleanse the resource.
func (p *PodDisruptionBudget) Sanitize(ctx context.Context) error {
	for fqn, pdb := range p.ListPodDisruptionBudgets() {
		p.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		p.checkInUse(ctx, pdb)

		if p.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			p.ClearOutcome(fqn)
		}
	}

	return nil
}

func (p *PodDisruptionBudget) checkInUse(ctx context.Context, pdb *pv1beta1.PodDisruptionBudget) {
	m, err := metav1.LabelSelectorAsMap(pdb.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msg("No selectors found")
		return
	}
	if p.GetPod(pdb.Namespace, m) == nil {
		p.AddCode(ctx, 900)
		return
	}
}
