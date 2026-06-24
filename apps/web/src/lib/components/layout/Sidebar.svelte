<script lang="ts">
	interface Props {
		collapsed?: boolean;
	}

	let { collapsed = false }: Props = $props();
</script>

<aside class="sidebar" class:collapsed aria-label="Campaign navigation">
	<div class="sidebar-header">
		<h2>Campaigns</h2>
	</div>
	<nav class="sidebar-nav" aria-label="Campaign pages">
		<!-- Placeholder: campaign tree will be populated in Phase 1 -->
		<div class="placeholder-item">
			<span class="placeholder-icon">📁</span>
			<span class="placeholder-text">No campaigns yet</span>
		</div>
		<div class="placeholder-hint">
			Create your first campaign to get started.
		</div>
	</nav>
</aside>

<style>
	.sidebar {
		width: var(--sidebar-width);
		height: 100%;
		background: var(--color-surface);
		border-right: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		flex-shrink: 0;
		transition: width 0.2s ease;
	}

	.sidebar.collapsed {
		width: 0;
		overflow: hidden;
		border-right: none;
	}

	.sidebar-header {
		padding: 1rem;
		border-bottom: 1px solid var(--color-border);
	}

	.sidebar-header h2 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.sidebar-nav {
		flex: 1;
		padding: 0.75rem;
	}

	.placeholder-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		border-radius: 6px;
		color: var(--color-text-muted);
		font-size: 0.875rem;
	}

	.placeholder-icon {
		font-size: 1rem;
	}

	.placeholder-text {
		opacity: 0.7;
	}

	.placeholder-hint {
		padding: 0.5rem 0.75rem;
		font-size: 0.75rem;
		color: var(--color-text-muted);
		opacity: 0.5;
		font-style: italic;
	}

	@media (max-width: 768px) {
		.sidebar {
			position: fixed;
			left: 0;
			top: var(--header-height);
			bottom: var(--comment-bar-height);
			z-index: 100;
			box-shadow: 2px 0 8px rgba(0, 0, 0, 0.3);
		}

		.sidebar:not(.collapsed) {
			width: var(--sidebar-width);
		}
	}
</style>
