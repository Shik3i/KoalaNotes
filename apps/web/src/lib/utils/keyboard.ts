type ShortcutHandler = (e: KeyboardEvent) => void;

interface Shortcut {
	key: string;
	ctrl?: boolean;
	meta?: boolean;
	shift?: boolean;
	handler: ShortcutHandler;
	description: string;
}

let shortcuts: Shortcut[] = [];
let listener: ((e: KeyboardEvent) => void) | null = null;

function isMatch(e: KeyboardEvent, s: Shortcut): boolean {
	if (e.key.toLowerCase() !== s.key.toLowerCase()) return false;
	if (s.ctrl && !e.ctrlKey) return false;
	if (s.meta && !e.metaKey) return false;
	if (s.shift && !e.shiftKey) return false;
	// If shortcut requires ctrl/meta, ignore those modifier-only keys
	if (!s.ctrl && !s.meta) {
		if (e.ctrlKey || e.metaKey) return false;
	}
	return true;
}

function handleKeydown(e: KeyboardEvent) {
	// Ignore when typing in input/textarea
	const tag = (e.target as HTMLElement)?.tagName;
	if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return;

	for (const s of shortcuts) {
		if (isMatch(e, s)) {
			e.preventDefault();
			s.handler(e);
			return;
		}
	}
}

/** Register global keyboard shortcuts. Hot-swappable on re-call. */
export function registerShortcuts(list: Shortcut[]): () => void {
	if (typeof window !== 'undefined') {
		if (listener) window.removeEventListener('keydown', listener);
		listener = handleKeydown;
		window.addEventListener('keydown', listener);
	}
	shortcuts = list;
	let unregistered = false;
	return () => {
		if (unregistered) return;
		unregistered = true;
		if (listener) {
			window.removeEventListener('keydown', listener);
			listener = null;
		}
		shortcuts = [];
	};
}

export type { Shortcut };
