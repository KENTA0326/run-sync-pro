package handler

import (
	"context"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/usecase"
	"github.com/KENTA0326/run-sync-pro/model"
)

// --- handler 層テスト用の手書きモック（利用側ポートに対する fake 実装） ---

type fakeAuthUC struct {
	signUp func(ctx context.Context, input usecase.SignUpInput) (*usecase.SignUpOutput, error)
	login  func(ctx context.Context, input usecase.LoginInput) (*usecase.LoginOutput, error)
}

func (f *fakeAuthUC) SignUp(ctx context.Context, input usecase.SignUpInput) (*usecase.SignUpOutput, error) {
	if f.signUp != nil {
		return f.signUp(ctx, input)
	}
	return &usecase.SignUpOutput{UserID: 1}, nil
}

func (f *fakeAuthUC) Login(ctx context.Context, input usecase.LoginInput) (*usecase.LoginOutput, error) {
	if f.login != nil {
		return f.login(ctx, input)
	}
	return &usecase.LoginOutput{Token: "test-token"}, nil
}

type fakeShoeUC struct {
	create func(ctx context.Context, input usecase.CreateShoeInput) (*domain.Shoe, error)
	list   func(ctx context.Context, userID uint, limit, offset int) (*usecase.ShoeListOutput, error)
}

func (f *fakeShoeUC) Create(ctx context.Context, input usecase.CreateShoeInput) (*domain.Shoe, error) {
	if f.create != nil {
		return f.create(ctx, input)
	}
	return nil, nil
}

func (f *fakeShoeUC) GetByID(ctx context.Context, id, userID uint) (*domain.Shoe, error) {
	return nil, nil
}

func (f *fakeShoeUC) List(ctx context.Context, userID uint, limit, offset int) (*usecase.ShoeListOutput, error) {
	if f.list != nil {
		return f.list(ctx, userID, limit, offset)
	}
	return &usecase.ShoeListOutput{}, nil
}

func (f *fakeShoeUC) Delete(ctx context.Context, id, userID uint) error {
	return nil
}

type fakeTrainingLogUC struct{}

func (f *fakeTrainingLogUC) Create(ctx context.Context, input usecase.CreateTrainingLogInput) error {
	return nil
}
func (f *fakeTrainingLogUC) GetByID(ctx context.Context, id, userID uint) (*domain.TrainingLog, error) {
	return nil, nil
}
func (f *fakeTrainingLogUC) List(ctx context.Context, userID uint, limit, offset int) (*usecase.TrainingLogListOutput, error) {
	return nil, nil
}
func (f *fakeTrainingLogUC) ListAll(ctx context.Context, userID uint) ([]*domain.TrainingLog, error) {
	return nil, nil
}

type fakePasswordResetUC struct{}

func (f *fakePasswordResetUC) RequestReset(ctx context.Context, input usecase.RequestResetInput) (*usecase.RequestResetOutput, error) {
	return nil, nil
}
func (f *fakePasswordResetUC) ConfirmReset(ctx context.Context, input usecase.ConfirmResetInput) error {
	return nil
}

type fakeAnalyzer struct {
	analyze func(logs []model.TrainingLog) model.AnalysisResponse
}

func (f *fakeAnalyzer) AnalyzeByMonth(logs []model.TrainingLog) model.AnalysisResponse {
	if f.analyze != nil {
		return f.analyze(logs)
	}
	return model.AnalysisResponse{}
}

type fakeVDOT struct {
	vdot   float64
	paces  model.TrainingPaces
	riegel model.RacePredictions
}

func (f *fakeVDOT) CalculateVDOT(distanceMeters, timeSeconds float64) float64 {
	return f.vdot
}

func (f *fakeVDOT) CalculateTrainingPaces(vdot float64) model.TrainingPaces {
	return f.paces
}

func (f *fakeVDOT) CalculateRiegelPredictions(distanceMeters, timeSeconds, exponent float64) model.RacePredictions {
	return f.riegel
}

type fakeAnalysisCache struct {
	get        func(ctx context.Context, userID domain.UserID) (model.AnalysisResponse, bool, error)
	set        func(ctx context.Context, userID domain.UserID, res model.AnalysisResponse) error
	invalidate func(ctx context.Context, userID domain.UserID) error
}

func (f *fakeAnalysisCache) Get(ctx context.Context, userID domain.UserID) (model.AnalysisResponse, bool, error) {
	if f.get != nil {
		return f.get(ctx, userID)
	}
	return model.AnalysisResponse{}, false, nil
}

func (f *fakeAnalysisCache) Set(ctx context.Context, userID domain.UserID, res model.AnalysisResponse) error {
	if f.set != nil {
		return f.set(ctx, userID, res)
	}
	return nil
}

func (f *fakeAnalysisCache) Invalidate(ctx context.Context, userID domain.UserID) error {
	if f.invalidate != nil {
		return f.invalidate(ctx, userID)
	}
	return nil
}

func newTestHandlers(
	authUC AuthUseCasePort,
	shoeUC ShoeUseCasePort,
	analyzer TrainingAnalyzer,
	cache MonthlyAnalysisCache,
	vdot VDOTCalculator,
) *Handlers {
	return NewHandlers(
		nil,
		nil,
		authUC,
		shoeUC,
		&fakeTrainingLogUC{},
		&fakePasswordResetUC{},
		analyzer,
		cache,
		vdot,
		nil,
		func() time.Duration { return time.Hour },
		func() string { return "http://localhost:3001" },
	)
}
