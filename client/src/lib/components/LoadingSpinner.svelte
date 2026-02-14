<script lang="ts">
	interface Props {
		fullPage?: boolean;
	}

	let { fullPage = false }: Props = $props();
</script>

<div class="flex items-center justify-center {fullPage ? 'py-24' : 'py-12'}">
	<img src="/logo.png" alt="Savvy" class="h-16 animate-pulse" />
</div>
