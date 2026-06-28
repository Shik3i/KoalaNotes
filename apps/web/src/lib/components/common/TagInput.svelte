<script lang="ts">
	interface Props {
		tags: string[];
		onchange?: (tags: string[]) => void;
	}

	let { tags = $bindable(), onchange }: Props = $props();

	let input = $state('');

	function addTag(raw: string) {
		const tag = raw.trim().toLowerCase().replace(/[^a-z0-9_-]/g, '');
		if (!tag || tags.includes(tag)) return;
		tags = [...tags, tag];
		onchange?.(tags);
		input = '';
	}

	function removeTag(tag: string) {
		tags = tags.filter((t) => t !== tag);
		onchange?.(tags);
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' || e.key === ',') {
			e.preventDefault();
			addTag(input);
		}
		if (e.key === 'Backspace' && !input && tags.length > 0) {
			removeTag(tags[tags.length - 1]);
		}
	}

	function onPaste(e: ClipboardEvent) {
		const text = e.clipboardData?.getData('text') ?? '';
		if (/,|;|\n/.test(text)) {
			e.preventDefault();
			const parts = text.split(/[,;\n]+/).map((s) => s.trim()).filter(Boolean);
			for (const part of parts) {
				addTag(part);
			}
		}
	}
</script>

<div class="tag-input" role="listbox" aria-label="Tags">
	{#each tags as tag (tag)}
		<span class="tag-pill" role="option" aria-selected="true">
			<span>{tag}</span>
			<button
				class="tag-remove"
				onclick={() => removeTag(tag)}
				aria-label="Remove tag {tag}"
			>
				&times;
			</button>
		</span>
	{/each}
	<input
		type="text"
		bind:value={input}
		onkeydown={onKeydown}
		onpaste={onPaste}
		class="tag-field"
		placeholder={tags.length === 0 ? 'Add tags...' : ''}
		aria-label="Add a tag"
	/>
</div>

<style>
	.tag-input {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.375rem;
		padding: 0.375rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 6px;
		min-height: 36px;
		cursor: text;
		transition: border-color 0.15s;
	}

	.tag-input:focus-within {
		border-color: var(--color-primary);
	}

	.tag-pill {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.125rem 0.375rem;
		background: var(--color-surface-alt);
		border: 1px solid var(--color-border);
		border-radius: 4px;
		font-size: 0.75rem;
		color: var(--color-text-muted);
	}

	.tag-remove {
		background: none;
		border: none;
		color: var(--color-text-muted);
		font-size: 0.875rem;
		line-height: 1;
		padding: 0;
		cursor: pointer;
		opacity: 0.5;
		transition: opacity 0.15s;
	}

	.tag-remove:hover {
		opacity: 1;
		color: var(--color-danger);
	}

	.tag-field {
		flex: 1;
		min-width: 80px;
		border: none;
		outline: none;
		background: transparent;
		color: var(--color-text);
		font-size: 0.8125rem;
		font-family: inherit;
		padding: 0.125rem 0;
	}

	.tag-field::placeholder {
		color: var(--color-text-muted);
		opacity: 0.4;
	}
</style>
