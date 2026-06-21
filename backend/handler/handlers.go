package handler

import "gorm.io/gorm"

// Handlers は HTTP ハンドラとその依存（DB・ドメインサービス）をまとめる。
type Handlers struct {
	db               *gorm.DB
	auth             Auth
	authUC           AuthUseCasePort
	shoeUC           ShoeUseCasePort
	trainingLogUC    TrainingLogUseCasePort
	passwordResetUC  PasswordResetUseCasePort
	analyzer         TrainingAnalyzer
	analysisCache    MonthlyAnalysisCache
	vdot             VDOTCalculator
	splits           MarathonSplits
	passwordResetTTL PasswordResetTTLFunc
	frontendBaseURL  FrontendBaseURLFunc
}

// NewHandlers は依存を明示的に渡して構築する（main またはテストで具体実装・モックを注入する）。
func NewHandlers(
	db *gorm.DB,
	auth Auth,
	authUC AuthUseCasePort,
	shoeUC ShoeUseCasePort,
	trainingLogUC TrainingLogUseCasePort,
	passwordResetUC PasswordResetUseCasePort,
	analyzer TrainingAnalyzer,
	analysisCache MonthlyAnalysisCache,
	vdot VDOTCalculator,
	splits MarathonSplits,
	passwordResetTTL PasswordResetTTLFunc,
	frontendBaseURL FrontendBaseURLFunc,
) *Handlers {
	return &Handlers{
		db:               db,
		auth:             auth,
		authUC:           authUC,
		shoeUC:           shoeUC,
		trainingLogUC:    trainingLogUC,
		passwordResetUC:  passwordResetUC,
		analyzer:         analyzer,
		analysisCache:    analysisCache,
		vdot:             vdot,
		splits:           splits,
		passwordResetTTL: passwordResetTTL,
		frontendBaseURL:  frontendBaseURL,
	}
}

// DB はルーティング層でミドルウェアに注入するための参照を返す。
func (h *Handlers) DB() *gorm.DB {
	return h.db
}
