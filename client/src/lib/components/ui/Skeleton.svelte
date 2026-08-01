<script lang="ts">
	// Loading placeholder block. Shape (height/width/rounding) comes from the
	// call site via `class`; the component only carries the shimmer + neutral
	// fill so pulses look identical everywhere.
	let { class: klass = '' }: { class?: string } = $props();
</script>

<div class="animate-pulse rounded bg-border-soft {klass}"></div>
