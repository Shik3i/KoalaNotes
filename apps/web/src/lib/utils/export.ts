import type { Note } from '$lib/types/models';

/** Trigger a file download in the browser. */
function download(filename: string, content: string, mimeType = 'text/markdown') {
	const blob = new Blob([content], { type: mimeType });
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}

/** Format a note as a Markdown document with frontmatter-style header. */
export function formatNoteAsMarkdown(note: Note): string {
	const lines: string[] = [];

	// Title
	lines.push(`# ${note.title}`);
	lines.push('');

	// Metadata
	if (note.template_type && note.template_type !== 'blank') {
		lines.push(`> **Template**: ${note.template_type}`);
	}
	if (note.tags.length > 0) {
		lines.push(`> **Tags**: ${note.tags.join(', ')}`);
	}
	if (note.updated_at) {
		lines.push(`> **Last updated**: ${new Date(note.updated_at).toLocaleString()}`);
	}
	if (lines.length > 1) lines.push('');

	// Content
	lines.push(note.content || '*No content.*');
	lines.push('');

	return lines.join('\n');
}

/** Format all notes as a bundle Markdown document. */
export function formatNotesAsMarkdown(notes: Note[], campaignName: string): string {
	const lines: string[] = [];

	lines.push(`# ${campaignName} — All Notes`);
	lines.push('');
	lines.push(`> Exported on ${new Date().toLocaleString()}`);
	lines.push(`> ${notes.length} note${notes.length === 1 ? '' : 's'}`);
	lines.push('');
	lines.push('---');
	lines.push('');

	for (let i = 0; i < notes.length; i++) {
		lines.push(formatNoteAsMarkdown(notes[i]));
		if (i < notes.length - 1) {
			lines.push('---');
			lines.push('');
		}
	}

	return lines.join('\n');
}

/** Export a single note as a downloadable .md file. */
export function exportNoteAsMarkdown(note: Note): void {
	const filename = `${note.title.replace(/[^a-zA-Z0-9_-]/g, '_')}.md`;
	download(filename, formatNoteAsMarkdown(note));
}

/** Export all notes as a downloadable bundle .md file. */
export function exportNotesAsMarkdown(notes: Note[], campaignName: string): void {
	const safeName = campaignName.replace(/[^a-zA-Z0-9_-]/g, '_');
	const filename = `${safeName}_notes.md`;
	download(filename, formatNotesAsMarkdown(notes, campaignName));
}

/** Format elapsed seconds as HH:MM:SS. */
export function formatElapsed(seconds: number): string {
	if (!Number.isFinite(seconds) || seconds < 0) return '00:00:00';
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = seconds % 60;
	return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}
