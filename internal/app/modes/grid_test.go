package modes

import (
	"context"
	"image"
	"testing"

	"go.uber.org/zap"

	"github.com/y3owk1n/neru/internal/app/components"
	gridcomponent "github.com/y3owk1n/neru/internal/app/components/grid"
	scrollcomponent "github.com/y3owk1n/neru/internal/app/components/scroll"
	"github.com/y3owk1n/neru/internal/app/services"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/domain"
	domainGrid "github.com/y3owk1n/neru/internal/domain/grid"
	"github.com/y3owk1n/neru/internal/domain/modecmd"
	"github.com/y3owk1n/neru/internal/domain/state"
	portmocks "github.com/y3owk1n/neru/internal/ports/mocks"
)

func TestHandleGridModeKey_CompleteSelectionDoesNotMoveWhenCursorFollowSelectionDisabled(
	t *testing.T,
) {
	moveCount := 0

	gridInstance := domainGrid.NewGridWithLabels(
		"ABCD",
		"",
		"",
		image.Rect(0, 0, 100, 100),
		zap.NewNop(),
	)

	manager := domainGrid.NewManager(
		gridInstance,
		domain.GridDimensions{Rows: 3, Cols: 3},
		"asdfghjkl",
		nil,
		nil,
		zap.NewNop(),
	)

	handler := newHandlerWithState(handlerState{
		config: &config.Config{
			Grid: config.GridConfig{
				Enabled:    true,
				Characters: "ABCD",
				Hotkeys:    map[string]config.StringOrStringArray{},
			},
		},
		logger: zap.NewNop(),
		actionService: services.NewActionService(
			&portmocks.MockAccessibilityPort{},
			&portmocks.MockOverlayPort{},
			&portmocks.MockSystemPort{
				MoveCursorToPointFunc: func(_ context.Context, _ image.Point, _ bool) error {
					moveCount++

					return nil
				},
			},
			zap.NewNop(),
		),
		grid: &components.GridComponent{
			Manager: manager,
			Router:  domainGrid.NewRouter(manager, zap.NewNop()),
			Context: &gridcomponent.Context{},
		},
		screenBounds: image.Rect(0, 0, 100, 100),
	})

	handler.grid.Context.SetCursorFollowSelection(false)

	handler.handleGridModeKey("A")
	handler.handleGridModeKey("A")
	handler.handleGridModeKey("A")

	if moveCount != 0 {
		t.Fatalf("handleGridModeKey() moved cursor %d times, want 0", moveCount)
	}

	if _, ok := handler.grid.Context.SelectionPoint(); !ok {
		t.Fatal("expected final selection point to be stored")
	}
}

func TestHandleGridModeKey_EnteringSubgridDoesNotMoveWhenCursorFollowSelectionDisabled(
	t *testing.T,
) {
	moveCount := 0

	gridInstance := domainGrid.NewGridWithLabels(
		"ABCD",
		"",
		"",
		image.Rect(0, 0, 100, 100),
		zap.NewNop(),
	)

	handler := newHandlerWithState(handlerState{
		config: &config.Config{
			Grid: config.GridConfig{
				Enabled:    true,
				Characters: "ABCD",
				Hotkeys:    map[string]config.StringOrStringArray{},
			},
		},
		logger: zap.NewNop(),
		actionService: services.NewActionService(
			&portmocks.MockAccessibilityPort{},
			&portmocks.MockOverlayPort{},
			&portmocks.MockSystemPort{
				MoveCursorToPointFunc: func(_ context.Context, _ image.Point, _ bool) error {
					moveCount++

					return nil
				},
			},
			zap.NewNop(),
		),
		grid: &components.GridComponent{
			Context: &gridcomponent.Context{},
		},
		overlayPort:  &portmocks.MockOverlayPort{},
		screenBounds: image.Rect(0, 0, 100, 100),
	})

	handler.initializeGridManager(gridInstance)
	handler.grid.Router = domainGrid.NewRouter(handler.grid.Manager, zap.NewNop())
	handler.grid.Context.SetCursorFollowSelection(false)

	handler.handleGridModeKey("A")
	handler.handleGridModeKey("A")

	if moveCount != 0 {
		t.Fatalf(
			"handleGridModeKey() moved cursor %d times while entering subgrid, want 0",
			moveCount,
		)
	}

	if _, ok := handler.grid.Context.SelectionPoint(); !ok {
		t.Fatal("expected subgrid entry selection point to be stored")
	}
}

// TestApplyGridFlags_TellsAbsentOnExitFromEmptyOne pins the --on-exit contract
// for grid, where nil and empty are different values rather than two spellings
// of nothing. A repeat re-activation carries no --on-exit and must keep the
// steps the mode was activated with; a command that gave --on-exit no steps is
// asking for none to run.
func TestApplyGridFlags_TellsAbsentOnExitFromEmptyOne(t *testing.T) {
	stored := []string{"exec done"}

	tests := []struct {
		name      string
		onExit    []string
		isRefresh bool
		want      int
	}{
		{name: "absent on a refresh keeps the stored steps", isRefresh: true, want: 1},
		{
			name:      "given but empty on a refresh clears them",
			onExit:    []string{},
			isRefresh: true,
			want:      0,
		},
		{name: "absent on a fresh activation clears them", want: 0},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := &gridcomponent.Context{}
			ctx.SetOnExit(stored)

			applyGridFlags(ctx, modecmd.Activation{OnExit: testCase.onExit}, testCase.isRefresh)

			if len(ctx.OnExit()) != testCase.want {
				t.Errorf("OnExit = %v, want %d step(s)", ctx.OnExit(), testCase.want)
			}
		})
	}
}

func TestHandleGridModeKey_OnSelectDispatchesAndExits(t *testing.T) {
	moveCount := 0
	gotSteps := make(chan []string, 1)

	gridInstance := domainGrid.NewGridWithLabels(
		"ABCD",
		"",
		"",
		image.Rect(0, 0, 100, 100),
		zap.NewNop(),
	)
	manager := domainGrid.NewManager(
		gridInstance,
		domain.GridDimensions{Rows: 3, Cols: 3},
		"asdfghjkl",
		nil,
		nil,
		zap.NewNop(),
	)

	appState := state.NewAppState()
	appState.SetMode(domain.ModeGrid)

	handler := newHandlerWithState(handlerState{
		appState:    appState,
		cursorState: state.NewCursorState(),
		config: &config.Config{
			Grid: config.GridConfig{
				Enabled:    true,
				Characters: "ABCD",
				Hotkeys:    map[string]config.StringOrStringArray{},
				OnSelect:   config.StringOrStringArray{"scroll"},
			},
		},
		logger: zap.NewNop(),
		actionService: services.NewActionService(
			&portmocks.MockAccessibilityPort{},
			&portmocks.MockOverlayPort{},
			&portmocks.MockSystemPort{
				MoveCursorToPointFunc: func(_ context.Context, _ image.Point, _ bool) error {
					moveCount++

					return nil
				},
			},
			zap.NewNop(),
		),
		grid: &components.GridComponent{
			Manager: manager,
			Router:  domainGrid.NewRouter(manager, zap.NewNop()),
			Context: &gridcomponent.Context{},
		},
		scroll: &components.ScrollComponent{
			Context: &scrollcomponent.Context{},
		},
		screenBounds: image.Rect(0, 0, 100, 100),
		executeActionSequence: func(source string, steps []string) {
			if source == "on-select" {
				gotSteps <- steps
			}
		},
		overlayPort: &portmocks.MockOverlayPort{},
	})

	handler.handleGridModeKey("A")
	handler.handleGridModeKey("A")
	handler.handleGridModeKey("A")

	if moveCount != 1 {
		t.Fatalf("handleGridModeKey() moved cursor %d times, want 1", moveCount)
	}

	// Single mode transition should immediately switch to scroll mode without exiting to idle
	if appState.CurrentMode() != domain.ModeScroll {
		t.Fatalf("expected mode to transition to scroll, got %v", appState.CurrentMode())
	}

	if !handler.scroll.Context.IsActive() {
		t.Fatal("expected scroll context to be active")
	}
}
