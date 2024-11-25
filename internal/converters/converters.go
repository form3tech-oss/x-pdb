package converters

import (
	statepb "github.com/form3tech-oss/x-pdb/pkg/protos/state"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func ConvertLabelSelectorToMetaV1(s *statepb.LabelSelector) *metav1.LabelSelector {
	mappedMatchExpressions := make([]metav1.LabelSelectorRequirement, len(s.MatchExpressions))

	for i := range s.MatchExpressions {
		mappedMatchExpressions[i] = ConvertLabelSelectorRequirementToMetaV1(s.MatchExpressions[i])
	}

	return &metav1.LabelSelector{
		MatchLabels:      s.MatchLabels,
		MatchExpressions: mappedMatchExpressions,
	}
}

func ConvertLabelSelectorRequirementToMetaV1(r *statepb.LabelSelectorRequirement) metav1.LabelSelectorRequirement {
	req := metav1.LabelSelectorRequirement{
		Values: r.Values,
	}

	if r.Key != nil {
		req.Key = *r.Key
	}

	if r.Operator != nil {
		req.Operator = metav1.LabelSelectorOperator(*r.Operator)
	}

	return req
}

func ConvertLabelSelectorToState(s *metav1.LabelSelector) *statepb.LabelSelector {
	mappedMatchExpressions := make([]*statepb.LabelSelectorRequirement, len(s.MatchExpressions))

	for i := range s.MatchExpressions {
		mappedMatchExpressions[i] = ConvertLabelSelectorRequirementToState(&s.MatchExpressions[i])
	}

	return &statepb.LabelSelector{
		MatchLabels:      s.MatchLabels,
		MatchExpressions: mappedMatchExpressions,
	}
}

func ConvertLabelSelectorRequirementToState(r *metav1.LabelSelectorRequirement) *statepb.LabelSelectorRequirement {
	req := &statepb.LabelSelectorRequirement{
		Values:   r.Values,
		Key:      ptr.To(r.Key),
		Operator: ptr.To(string(r.Operator)),
	}

	return req
}
