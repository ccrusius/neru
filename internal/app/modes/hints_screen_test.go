package modes

import (
	"context"
	"image"
	"testing"

	portmocks "github.com/y3owk1n/neru/internal/ports/mocks"
)

func TestScreenBoundsForFocusedWindow_FallsBackWhenNilSystem(t *testing.T) {
	fallback := image.Rect(0, 0, 1920, 1080)
	got := screenBoundsForFocusedWindow(context.Background(), nil, fallback)
	if got != fallback {
		t.Errorf("screenBoundsForFocusedWindow(nil) = %v, want fallback %v", got, fallback)
	}
}

func TestScreenBoundsForFocusedWindow_FallsBackWhenNoWindow(t *testing.T) {
	fallback := image.Rect(0, 0, 1920, 1080)
	mockSystem := &portmocks.MockSystemPort{
		FocusedWindowBoundsFunc: func(context.Context) (image.Rectangle, bool, error) {
			return image.Rectangle{}, false, nil
		},
	}

	got := screenBoundsForFocusedWindow(context.Background(), mockSystem, fallback)
	if got != fallback {
		t.Errorf("screenBoundsForFocusedWindow(no window) = %v, want fallback %v", got, fallback)
	}
}

func TestScreenBoundsForFocusedWindow_FastPathsWhenWindowInFallback(t *testing.T) {
	fallback := image.Rect(0, 0, 1920, 1080)
	mockSystem := &portmocks.MockSystemPort{
		FocusedWindowBoundsFunc: func(context.Context) (image.Rectangle, bool, error) {
			return image.Rect(100, 100, 500, 400), true, nil
		},
		ScreenNamesFunc: func(context.Context) ([]string, error) {
			return []string{"primary", "auxiliary"}, nil
		},
		ScreenBoundsByNameFunc: func(_ context.Context, name string) (image.Rectangle, bool, error) {
			if name == "auxiliary" {
				t.Fatal("auxiliary screen should not be queried when window center is already in fallback")
			}
			return fallback, true, nil
		},
	}

	got := screenBoundsForFocusedWindow(context.Background(), mockSystem, fallback)
	if got != fallback {
		t.Errorf("screenBoundsForFocusedWindow(in fallback) = %v, want fallback %v", got, fallback)
	}
}

func TestScreenBoundsForFocusedWindow_ResolvesAuxiliaryMonitor(t *testing.T) {
	primary := image.Rect(1920, 0, 5760, 2160)
	auxScreen := image.Rect(0, 0, 1920, 1200)

	mockSystem := &portmocks.MockSystemPort{
		FocusedWindowBoundsFunc: func(context.Context) (image.Rectangle, bool, error) {
			// Focused window on auxiliary screen
			return image.Rect(10, 30, 1910, 1180), true, nil
		},
		ScreenNamesFunc: func(context.Context) ([]string, error) {
			return []string{"DP-1", "eDP-1"}, nil
		},
		ScreenBoundsByNameFunc: func(_ context.Context, name string) (image.Rectangle, bool, error) {
			if name == "DP-1" {
				return primary, true, nil
			}
			if name == "eDP-1" {
				return auxScreen, true, nil
			}
			return image.Rectangle{}, false, nil
		},
	}

	got := screenBoundsForFocusedWindow(context.Background(), mockSystem, primary)
	if got != auxScreen {
		t.Errorf("screenBoundsForFocusedWindow() = %v, want auxScreen %v", got, auxScreen)
	}
}

func TestScreenBoundsForFocusedWindow_FallsBackWhenScreenNotFound(t *testing.T) {
	primary := image.Rect(1920, 0, 5760, 2160)

	mockSystem := &portmocks.MockSystemPort{
		FocusedWindowBoundsFunc: func(context.Context) (image.Rectangle, bool, error) {
			// Window off on an unknown coordinate
			return image.Rect(99999, 99999, 100000, 100000), true, nil
		},
		ScreenNamesFunc: func(context.Context) ([]string, error) {
			return []string{"DP-1"}, nil
		},
		ScreenBoundsByNameFunc: func(_ context.Context, name string) (image.Rectangle, bool, error) {
			return primary, true, nil
		},
	}

	got := screenBoundsForFocusedWindow(context.Background(), mockSystem, primary)
	if got != primary {
		t.Errorf("screenBoundsForFocusedWindow() = %v, want fallback %v", got, primary)
	}
}

func TestResolveHintsScreenBounds_UsesFocusedWindow(t *testing.T) {
	primary := image.Rect(1920, 0, 5760, 2160)
	auxScreen := image.Rect(0, 0, 1920, 1200)

	mockSystem := &portmocks.MockSystemPort{
		ScreenBoundsFunc: func(context.Context) (image.Rectangle, error) {
			return primary, nil
		},
		FocusedWindowBoundsFunc: func(context.Context) (image.Rectangle, bool, error) {
			return image.Rect(10, 30, 1910, 1180), true, nil
		},
		ScreenNamesFunc: func(context.Context) ([]string, error) {
			return []string{"DP-1", "eDP-1"}, nil
		},
		ScreenBoundsByNameFunc: func(_ context.Context, name string) (image.Rectangle, bool, error) {
			if name == "DP-1" {
				return primary, true, nil
			}
			if name == "eDP-1" {
				return auxScreen, true, nil
			}
			return image.Rectangle{}, false, nil
		},
	}

	h := &handlerState{
		system: mockSystem,
	}

	got := h.resolveHintsScreenBounds(context.Background())
	if got != auxScreen {
		t.Errorf("resolveHintsScreenBounds() = %v, want auxScreen %v", got, auxScreen)
	}
}


