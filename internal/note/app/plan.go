package app

type PlanTier uint8

const (
	PlanTierUnspecified PlanTier = iota
	PlanTierFree
	PlanTierPlus
	PlanTierPro
)

func (t PlanTier) IsValid() bool {
	switch t {
	case PlanTierUnspecified, PlanTierFree, PlanTierPlus, PlanTierPro:
		return true
	default:
		return false
	}
}

func (t PlanTier) String() string {
	switch t {
	case PlanTierFree:
		return "free"
	case PlanTierPlus:
		return "plus"
	case PlanTierPro:
		return "pro"
	case PlanTierUnspecified:
		return "unspecified"
	default:
		return "unknown"
	}
}

type PlanFeature uint8

const (
	PlanFeatureUnspecified PlanFeature = iota
	PlanFeatureWorkspaceGraph
)

func (f PlanFeature) IsValid() bool {
	switch f {
	case PlanFeatureUnspecified, PlanFeatureWorkspaceGraph:
		return true
	default:
		return false
	}
}

func (f PlanFeature) String() string {
	switch f {
	case PlanFeatureWorkspaceGraph:
		return "workspace_graph"
	case PlanFeatureUnspecified:
		return "unspecified"
	default:
		return "unknown"
	}
}

func ParsePlanFeature(s string) (PlanFeature, bool) {
	switch s {
	case PlanFeatureWorkspaceGraph.String():
		return PlanFeatureWorkspaceGraph, true
	default:
		return PlanFeatureUnspecified, false
	}
}

type PlanLimit uint8

const (
	PlanLimitUnspecified PlanLimit = iota
	PlanLimitMembers
	PlanLimitNoteGraphMaxDepth
	PlanLimitTrashRetentionDays
)

func (l PlanLimit) IsValid() bool {
	switch l {
	case PlanLimitUnspecified,
		PlanLimitMembers,
		PlanLimitNoteGraphMaxDepth,
		PlanLimitTrashRetentionDays:
		return true
	default:
		return false
	}
}

func (l PlanLimit) String() string {
	switch l {
	case PlanLimitMembers:
		return "members"
	case PlanLimitNoteGraphMaxDepth:
		return "note_graph_max_depth"
	case PlanLimitTrashRetentionDays:
		return "trash_retention_days"
	case PlanLimitUnspecified:
		return "unspecified"
	default:
		return "unknown"
	}
}

func ParsePlanLimit(s string) (PlanLimit, bool) {
	switch s {
	case PlanLimitMembers.String():
		return PlanLimitMembers, true
	case PlanLimitNoteGraphMaxDepth.String():
		return PlanLimitNoteGraphMaxDepth, true
	case PlanLimitTrashRetentionDays.String():
		return PlanLimitTrashRetentionDays, true
	default:
		return PlanLimitUnspecified, false
	}
}

type PlanQuota int64

const PlanQuotaUnlimited PlanQuota = -1

func (q PlanQuota) IsUnlimited() bool { return q == PlanQuotaUnlimited }

func (q PlanQuota) Allows(n int64) bool { return q.IsUnlimited() || n <= int64(q) }

type Entitlements struct {
	Plan     PlanTier
	Features map[PlanFeature]bool
	Limits   map[PlanLimit]PlanQuota
}

func (e Entitlements) Feature(f PlanFeature) bool { return e.Features[f] }

func (e Entitlements) Limit(l PlanLimit) PlanQuota { return e.Limits[l] }

type PlanEntitlements struct {
	Tier     PlanTier
	Features map[PlanFeature]bool
	Limits   map[PlanLimit]PlanQuota
}

var paidTiers = []PlanTier{PlanTierPlus, PlanTierPro}

func planByTier(plans []PlanEntitlements, tier PlanTier) (PlanEntitlements, bool) {
	for _, p := range plans {
		if p.Tier == tier {
			return p, true
		}
	}
	return PlanEntitlements{}, false
}

func RequiredPlanForFeature(plans []PlanEntitlements, f PlanFeature) PlanTier {
	for _, tier := range paidTiers {
		if p, ok := planByTier(plans, tier); ok && p.Features[f] {
			return tier
		}
	}
	return PlanTierPro
}

func RequiredPlanForLimit(plans []PlanEntitlements, l PlanLimit, needed int64) PlanTier {
	for _, tier := range paidTiers {
		if p, ok := planByTier(plans, tier); ok && p.Limits[l].Allows(needed) {
			return tier
		}
	}
	return PlanTierPro
}
