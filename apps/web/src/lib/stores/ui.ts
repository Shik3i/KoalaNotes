/** UI state stores. Currently placeholder - will be wired to real data in Phase 1. */

// Placeholder: session timer state
// Will use Svelte stores ($state/$derived) when wired up
export interface UIState {
	sidebarOpen: boolean;
	timelineOpen: boolean;
	sessionActive: boolean;
	sessionElapsed: number; // seconds
}

export const defaultUIState: UIState = {
	sidebarOpen: true,
	timelineOpen: false,
	sessionActive: false,
	sessionElapsed: 0
};
