package modes

import (
	"context"
	"image"
	"testing"

	"github.com/y3owk1n/neru/internal/ports"
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
		ScreensFunc: func(context.Context) ([]ports.Screen, error) {
			t.Fatal("Screens should not be queried when window center is already in fallback")
			return nil, nil
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
		ScreensFunc: func(context.Context) ([]ports.Screen, error) {
			return []ports.Screen{
				{Name: "DP-1", Bounds: primary},
				{Name: "eDP-1", Bounds: auxScreen},
			}, nil
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
		ScreensFunc: func(context.Context) ([]ports.Screen, error) {
			return []ports.Screen{
				{Name: "DP-1", Bounds: primary},
			}, nil
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
		ScreensFunc: func(context.Context) ([]ports.Screen, error) {
			return []ports.Screen{
				{Name: "DP-1", Bounds: primary},
				{Name: "eDP-1", Bounds: auxScreen},
			}, nil
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
