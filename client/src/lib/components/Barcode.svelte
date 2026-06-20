<script lang="ts">
	import { onMount } from 'svelte';
	import bwipjs from 'bwip-js';
	import { logger } from '$lib/utils/logger';

	const componentLogger = logger.child('Barcode');

	interface Props {
		value: string;
		type?: string;
		width?: number;
		height?: number;
		displayValue?: boolean;
		maxHeight?: number;
	}

	let {
		value,
		type = 'CODE128',
		width = 2,
		height = 50,
		displayValue = false,
		maxHeight
	}: Props = $props();

	let canvas = $state<HTMLCanvasElement>();
	let containerDiv = $state<HTMLDivElement>();
	let barcodeSection = $state<HTMLDivElement>();
	let hasError = $state(false);

	// Map barcode types to bwip-js BCID (Barcode ID)
	const formatMap: Record<string, string> = {
		// 1D Barcodes
		CODE128: 'code128',
		CODE39: 'code39',
		CODE93: 'code93',
		EAN13: 'ean13',
		EAN8: 'ean8',
		UPC: 'upca',
		UPCA: 'upca',
		UPCE: 'upce',
		ITF: 'interleaved2of5',
		ITF14: 'itf14',
		MSI: 'msi',
		CODABAR: 'codabar',
		Pharmacode: 'pharmacode',
		// ISBN/ISSN
		ISBN13: 'ean13', // ISBN-13 uses EAN-13 format
		ISBN10: 'isbn', // ISBN-10 (bwip-js: 'isbn')
		ISBN: 'isbn', // Alias for ISBN-10
		ISSN: 'issn', // International Standard Serial Number
		// 2D Barcodes
		QR: 'qrcode',
		QRCODE: 'qrcode',
		PDF417: 'pdf417',
		DATAMATRIX: 'datamatrix',
		AZTEC: 'azteccode',
		MAXICODE: 'maxicode'
	};

	onMount(() => {
		generateBarcode();
	});

	function generateBarcode() {
		if (!canvas || !value) return;

		const bcid = formatMap[type.toUpperCase()] || 'code128';

		// Check if this is a 2D barcode
		const is2D = [
			'qrcode',
			'pdf417',
			'datamatrix',
			'azteccode',
			'maxicode'
		].includes(bcid);

		// bwip-js options
		const options: Parameters<typeof bwipjs.toCanvas>[1] = {
			bcid,
			text: value,
			includetext: displayValue,
			textxalign: 'center'
		};

		// Different scaling for 1D vs 2D barcodes
		if (is2D) {
			options.scale = width || 2;
			// Long payloads (URLs, long tokens) produce dense QR codes whose
			// modules become too small to scan. Bump the module scale so each
			// module stays large enough for a phone camera. (Module size is set
			// by `scale`; the error-correction level only affects module count.)
			if (bcid === 'qrcode' && value.length > 100) {
				options.scale = Math.max(options.scale, 4);
			}
		} else {
			options.scale = width || 2;
			options.height = Math.max(10, height / 5);
		}

		try {
			bwipjs.toCanvas(canvas, options);
			hasError = false;
		} catch (err) {
			// If a type-specific format failed, fall back to CODE128 (accepts any content)
			if (bcid !== 'code128') {
				componentLogger.warn(
					`Barcode type "${type}" failed, falling back to CODE128:`,
					err
				);
				try {
					bwipjs.toCanvas(canvas, { ...options, bcid: 'code128' });
					hasError = false;
					return;
				} catch {
					// CODE128 fallback also failed
				}
			}
			componentLogger.error('Barcode generation failed:', err);
			hasError = true;
		}
	}

	// Regenerate when value or type changes
	$effect(() => {
		if (value || type) {
			generateBarcode();
		}
	});
</script>

<div bind:this={containerDiv} class="barcode-container">
	{#if !hasError}
		<div bind:this={barcodeSection} class="barcode-wrapper">
			<canvas
				bind:this={canvas}
				class="barcode-canvas"
				style={maxHeight
					? `max-height: ${maxHeight}px; height: ${maxHeight}px;`
					: ''}
			></canvas>
		</div>
	{/if}
</div>

<style>
	.barcode-container {
		position: relative;
		display: inline-block;
	}

	.barcode-wrapper {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		width: 100%;
	}

	.barcode-canvas {
		max-width: 100%;
		height: auto;
		object-fit: contain;
	}
</style>
