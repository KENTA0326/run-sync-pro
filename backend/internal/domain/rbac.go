package domain

// Resource は RBAC の permissions.resource に対応する値オブジェクト。
type Resource string

// Action は RBAC の permissions.action に対応する値オブジェクト。
type Action string

const (
	ResourceUserProfile  Resource = "user_profile"
	ResourceShoe         Resource = "shoe"
	ResourceTrainingLog  Resource = "training_log"
	ResourceAnalysis     Resource = "analysis"
	ResourceWildcard     Resource = "*"

	ActionRead     Action = "read"
	ActionWrite    Action = "write"
	ActionWildcard Action = "*"
)

func (r Resource) String() string { return string(r) }

func (r Resource) IsValid() bool {
	switch r {
	case ResourceUserProfile, ResourceShoe, ResourceTrainingLog, ResourceAnalysis, ResourceWildcard:
		return true
	default:
		return false
	}
}

func (a Action) String() string { return string(a) }

func (a Action) IsValid() bool {
	switch a {
	case ActionRead, ActionWrite, ActionWildcard:
		return true
	default:
		return false
	}
}
