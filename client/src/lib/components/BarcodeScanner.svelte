<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { logger } from '$lib/utils/logger';
	import { BarcodeDetector } from 'barcode-detector/pure';
	import { tick } from 'svelte';

	const componentLogger = logger.child('BarcodeScanner');

	interface Props {
		open?: boolean;
		onscan?: (event: { barcode: string; format?: string }) => void;
		onerror?: (event: { message: string }) => void;
	}

	let { open = $bindable(false), onscan, onerror }: Props = $props();

	let videoElement = $state<HTMLVideoElement>();
	let canvasElement: HTMLCanvasElement | null = null;
	let canvasCtx: CanvasRenderingContext2D | null = null;
	let modalRef = $state<HTMLDivElement | null>(null);
	let previousFocus = $state<HTMLElement | null>(null);

	let isScanning = $state(false);
	let isInitializing = $state(false);
	let scanMessage = $state($t('common.scanPositionBarcode'));
	let mediaStream: MediaStream | null = null;
	let scannerReady = $state(false);
	let animationFrameId: number | null = null;

	let barcodeDetector: BarcodeDetector | null = null;

	let torchEnabled = $state(false);
	let torchSupported = $state(false);

	let showSuccess = $state(false);
	let validationWarning = $state<string | null>(null);

	let scanAttempts = $state(0);
	let lastScanTime = $state(0);
	let scanningFeedback = $state<'idle' | 'searching' | 'detecting' | 'found'>(
		'idle'
	);

	let showDebugPanel = $state(false);
	let debugLogs = $state<string[]>([]);

	function addDebugLog(message: string) {
		const timestamp = new Date().toLocaleTimeString();
		debugLogs = [`[${timestamp}] ${message}`, ...debugLogs].slice(0, 10);
	}

	function portal(node: HTMLElement) {
		node.style.margin = '0';
		document.body.appendChild(node);

		return {
			destroy() {
				if (node.parentNode) {
					node.parentNode.removeChild(node);
				}
			}
		};
	}

	$effect(() => {
		if (open && !isScanning) {
			tick().then(() => startScanning());
		} else if (!open && isScanning) {
			stopScanning();
		}
	});

	async function startScanning() {
		if (isScanning) return;

		try {
			isInitializing = true;
			isScanning = true;
			scanMessage = $t('common.scanInitializing');
			scanningFeedback = 'idle';
			scanAttempts = 0;

			await tick();

			mediaStream = await navigator.mediaDevices.getUserMedia({
				video: {
					facingMode: 'environment', // Use back camera
					width: { ideal: 1280 },
					height: { ideal: 720 }
				}
			});

			if (!videoElement) {
				throw new Error('Video element not found');
			}

			videoElement.srcObject = mediaStream;
			videoElement.setAttribute('playsinline', '');
			videoElement.setAttribute('webkit-playsinline', '');
			videoElement.setAttribute('muted', '');

			await new Promise<void>((resolve, reject) => {
				if (!videoElement) {
					reject(new Error('Video element lost'));
					return;
				}

				videoElement.onloadedmetadata = () => {
					videoElement
						?.play()
						.then(() => {
							setTimeout(() => resolve(), 500);
						})
						.catch(reject);
				};
			});

			componentLogger.info('Video ready:', {
				width: videoElement.videoWidth,
				height: videoElement.videoHeight,
				readyState: videoElement.readyState
			});
			addDebugLog(
				`Video ready: ${videoElement.videoWidth}x${videoElement.videoHeight} (state: ${videoElement.readyState})`
			);

			barcodeDetector = new BarcodeDetector({
				formats: [
					'aztec',
					'code_128',
					'code_39',
					'code_93',
					'codabar',
					'data_matrix',
					'ean_13',
					'ean_8',
					'itf',
					'pdf417',
					'qr_code',
					'upc_a',
					'upc_e'
				]
			});

			// Create offscreen canvas for frame capture (required for iOS Safari
			// where createImageBitmap from video elements is unsupported)
			canvasElement = document.createElement('canvas');
			canvasElement.width = videoElement.videoWidth || 1280;
			canvasElement.height = videoElement.videoHeight || 720;
			canvasCtx = canvasElement.getContext('2d', { willReadFrequently: true });

			scannerReady = true;
			componentLogger.info('Using BarcodeDetector polyfill');
			addDebugLog(
				`Using: BarcodeDetector (canvas ${canvasElement.width}x${canvasElement.height})`
			);

			isInitializing = false;
			scanMessage = $t('common.scanPositionBarcode');
			scanningFeedback = 'searching';

			startScanningLoop();
			checkTorchSupport();
		} catch (err) {
			componentLogger.error('Scanner error:', err);
			const errorMessage =
				err instanceof Error ? err.message : 'Kamera-Zugriff fehlgeschlagen';
			scanMessage = errorMessage;
			onerror?.({ message: errorMessage });

			if (
				errorMessage.includes('Permission') ||
				errorMessage.includes('NotAllowedError')
			) {
				scanMessage = $t('common.scanCameraPermissionDenied');
			} else if (errorMessage.includes('NotFoundError')) {
				scanMessage = $t('common.scanNoCameraFound');
			}

			isScanning = false;
			isInitializing = false;
		}
	}

	function startScanningLoop() {
		if (!isScanning || !videoElement) {
			return;
		}

		const scan = async () => {
			if (!isScanning || !videoElement) {
				return;
			}

			try {
				const now = Date.now();
				if (now - lastScanTime < 150) {
					animationFrameId = requestAnimationFrame(scan);
					return;
				}
				lastScanTime = now;
				scanAttempts++;

				updateScanningFeedback();

				if (barcodeDetector && canvasCtx && canvasElement) {
					// Draw current video frame to canvas (required for iOS Safari
					// where detect(videoElement) fails silently)
					canvasCtx.drawImage(
						videoElement,
						0,
						0,
						canvasElement.width,
						canvasElement.height
					);
					const barcodes = await barcodeDetector.detect(canvasElement);

					if (barcodes.length > 0) {
						const barcode = barcodes[0];
						handleBarcodeDetected(barcode.rawValue, barcode.format);
						return;
					}
				}

				if (scanAttempts % 50 === 0) {
					addDebugLog(`Scan attempt #${scanAttempts} - still scanning...`);
				}
			} catch (err) {
				componentLogger.debug('Scan iteration error:', err);
			}

			animationFrameId = requestAnimationFrame(scan);
		};

		animationFrameId = requestAnimationFrame(scan);
	}

	function handleBarcodeDetected(barcode: string, format: string) {
		// Sanitize: strip control characters
		const sanitized = barcode.replace(/[\x00-\x1F\x7F-\x9F]/g, '');
		if (sanitized.length === 0 || sanitized.length > 255) {
			componentLogger.warn('Invalid barcode rejected:', {
				length: sanitized.length,
				format
			});
			addDebugLog(`❌ Rejected: invalid barcode (length: ${sanitized.length})`);
			onerror?.({ message: $t('common.scanInvalidBarcode') });
			return;
		}

		componentLogger.info('Barcode detected:', { barcode: sanitized, format });
		addDebugLog(`✅ Detected: ${sanitized} (${format})`);

		const validation = validateFormat(sanitized, format);
		if (validation.warning) {
			validationWarning = validation.warning;
			componentLogger.warn('Format validation warning:', validation);
		}

		scanningFeedback = 'found';
		scanMessage = $t('common.scanFound');
		showSuccess = true;
		setTimeout(() => {
			showSuccess = false;
			validationWarning = null;
		}, 500);

		onscan?.({ barcode: sanitized, format: mapBarcodeFormat(format) });
		close();
	}

	function updateScanningFeedback() {
		if (scanAttempts < 20) {
			scanningFeedback = 'searching';
			scanMessage = $t('common.scanPositionBarcode');
		} else if (scanAttempts < 50) {
			scanningFeedback = 'detecting';
			scanMessage = $t('common.scanSearching');
		} else if (scanAttempts < 100) {
			if (scanAttempts % 20 === 0) {
				const tips = [
					$t('common.scanTipBetterLight'),
					$t('common.scanTipMoveCloser'),
					$t('common.scanTipAvoidReflections'),
					$t('common.scanTipKeepStraight')
				];
				const tipIndex = Math.floor(Math.random() * tips.length);
				scanMessage = tips[tipIndex];
			}
		} else {
			if (torchSupported && !torchEnabled && scanAttempts % 30 === 0) {
				scanMessage = $t('common.scanTipUseTorch');
			}
		}
	}

	async function checkTorchSupport() {
		try {
			if (!mediaStream) return;

			const track = mediaStream.getVideoTracks()[0];
			const capabilities = track.getCapabilities() as any;

			torchSupported = capabilities?.torch || false;
			componentLogger.info('Torch support:', { supported: torchSupported });
		} catch (err) {
			componentLogger.error('Failed to check torch support:', err);
			torchSupported = false;
		}
	}

	async function toggleTorch() {
		if (!mediaStream || !isScanning) return;

		try {
			const track = mediaStream.getVideoTracks()[0];
			await track.applyConstraints({
				advanced: [{ torch: !torchEnabled } as any]
			});
			torchEnabled = !torchEnabled;
			componentLogger.info('Torch toggled:', { enabled: torchEnabled });
		} catch (err) {
			componentLogger.error('Failed to toggle torch:', err);
			scanMessage = $t('common.scanTorchUnavailable');
			setTimeout(() => {
				scanMessage = $t('common.scanPositionBarcode');
			}, 2000);
		}
	}

	function mapBarcodeFormat(format: string): string {
		if (!format) {
			return 'CODE128';
		}

		const normalized = format.replace(/[-_/\s]/g, '').toUpperCase();

		const formatMap: Record<string, string> = {
			QRCODE: 'QR',
			QR: 'QR',
			CODE128: 'CODE128',
			CODE39: 'CODE39',
			CODE93: 'CODE93',
			CODABAR: 'CODABAR',
			EAN8: 'EAN8',
			EAN13: 'EAN13',
			UPCA: 'UPCA',
			UPCE: 'UPCE',
			ITF: 'ITF',
			PDF417: 'PDF417',
			DATAMATRIX: 'DATAMATRIX',
			AZTEC: 'AZTEC'
		};

		const mapped = formatMap[normalized];

		if (mapped) {
			return mapped;
		}

		componentLogger.warn('Unknown barcode format:', {
			original: format,
			normalized
		});
		return normalized;
	}

	function validateFormat(
		barcode: string,
		format: string
	): { valid: boolean; warning?: string } {
		if (format.includes('ean_13') && !/^\d{13}$/.test(barcode)) {
			return { valid: false, warning: 'EAN-13 should have 13 digits' };
		}

		if (format.includes('ean_8') && !/^\d{8}$/.test(barcode)) {
			return { valid: false, warning: 'EAN-8 should have 8 digits' };
		}

		if (format.includes('upc_a') && !/^\d{12}$/.test(barcode)) {
			return { valid: false, warning: 'UPC-A should have 12 digits' };
		}

		return { valid: true };
	}

	function stopScanning() {
		if (!isScanning) return;

		componentLogger.info('Stopping scanner...');

		if (animationFrameId !== null) {
			cancelAnimationFrame(animationFrameId);
			animationFrameId = null;
		}

		if (mediaStream) {
			mediaStream.getTracks().forEach((track) => {
				track.stop();
				componentLogger.debug('Stopped video track:', {
					kind: track.kind,
					label: track.label
				});
			});
			mediaStream = null;
		}

		if (videoElement) {
			videoElement.srcObject = null;
		}

		isScanning = false;
		isInitializing = false;
		torchEnabled = false;
		showSuccess = false;
		validationWarning = null;
		scanningFeedback = 'idle';
		scanAttempts = 0;
		barcodeDetector = null;
		canvasElement = null;
		canvasCtx = null;
		scannerReady = false;

		componentLogger.info('Scanner stopped and camera released');
	}

	async function close() {
		stopScanning();
		open = false;
	}

	$effect(() => {
		if (open && modalRef) {
			previousFocus = document.activeElement as HTMLElement;
			const closeButton = modalRef.querySelector('button');
			closeButton?.focus();

			return () => {
				previousFocus?.focus();
			};
		}
	});
</script>

{#if open}
	<div
		bind:this={modalRef}
		use:portal
		class="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center"
		role="dialog"
		aria-modal="true"
		aria-labelledby="scanner-title"
		aria-describedby="scanner-status"
		tabindex="-1"
		onclick={(e) => {
			if (e.target === e.currentTarget) {
				close();
			}
		}}
		onkeydown={(e) => {
			if (e.key === 'Escape') {
				e.preventDefault();
				close();
			}
		}}
	>
		<div class="bg-white rounded-lg scanner-container w-full m-4 p-4 md:p-6">
			<div class="flex justify-between items-center mb-4">
				<h3 id="scanner-title" class="text-lg font-semibold">
					{$t('common.scanningTitle')}
				</h3>
				<button
					type="button"
					onclick={close}
					class="text-gray-500 hover:text-gray-700 flex-shrink-0"
					aria-label={$t('common.close')}
				>
					<svg
						class="w-6 h-6"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						></path>
					</svg>
				</button>
			</div>

			<!-- Scanner Method Indicator -->
			{#if scannerReady}
				<div class="text-xs text-gray-500 mb-2 flex items-center gap-2">
					<div class="w-2 h-2 rounded-full bg-green-500"></div>
					BarcodeDetector
				</div>
			{/if}

			<div class="scanner-video-container">
				<!-- Video element for camera stream -->
				<video
					bind:this={videoElement}
					class="w-full rounded-lg"
					playsinline
					muted
				></video>

				<!-- Scanner overlay with guide -->
				<div class="scanner-overlay">
					<div
						class="scanner-guide"
						class:success={showSuccess}
						class:searching={scanningFeedback === 'searching'}
						class:detecting={scanningFeedback === 'detecting'}
						class:found={scanningFeedback === 'found'}
					></div>

					<!-- Loading spinner during initialization -->
					{#if isInitializing}
						<div class="scanner-loading">
							<svg
								class="animate-spin h-12 w-12 text-cyan-500"
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
							>
								<circle
									class="opacity-25"
									cx="12"
									cy="12"
									r="10"
									stroke="currentColor"
									stroke-width="4"
								></circle>
								<path
									class="opacity-75"
									fill="currentColor"
									d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
								></path>
							</svg>
						</div>
					{/if}

					<!-- Scan progress indicator -->
					{#if scanAttempts > 30 && !isInitializing}
						<div class="scan-progress">
							<div class="scan-progress-bar">
								<div
									class="scan-progress-fill"
									style="width: {Math.min((scanAttempts / 100) * 100, 100)}%"
								></div>
							</div>
						</div>
					{/if}
				</div>

				<!-- Torch/Flash Control -->
				{#if isScanning && torchSupported}
					<button
						type="button"
						onclick={toggleTorch}
						class="torch-button"
						class:active={torchEnabled}
						aria-label={torchEnabled
							? $t('common.scanTorchOff')
							: $t('common.scanTorchOn')}
						title={torchEnabled
							? $t('common.scanTorchOff')
							: $t('common.scanTorchOn')}
					>
						{#if torchEnabled}
							<!-- Flash On Icon -->
							<svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
								<path
									fill-rule="evenodd"
									d="M11.3 1.046A1 1 0 0112 2v5h4a1 1 0 01.82 1.573l-7 10A1 1 0 018 18v-5H4a1 1 0 01-.82-1.573l7-10a1 1 0 011.12-.38z"
									clip-rule="evenodd"
								></path>
							</svg>
						{:else}
							<!-- Flash Off Icon -->
							<svg
								class="w-6 h-6"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M13 10V3L4 14h7v7l9-11h-7z"
								></path>
							</svg>
						{/if}
					</button>
				{/if}
			</div>

			<!-- Status Messages -->
			<div class="mt-4 space-y-2">
				<p id="scanner-status" class="text-sm text-gray-600 text-center">
					{scanMessage}
				</p>
				{#if validationWarning}
					<p
						class="text-xs text-amber-600 text-center flex items-center justify-center gap-1"
					>
						<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
								clip-rule="evenodd"
							></path>
						</svg>
						{validationWarning}
					</p>
				{/if}
			</div>

			<!-- Debug Panel -->
			{#if showDebugPanel}
				<div class="debug-panel">
					<div class="debug-header">
						<span class="text-xs font-semibold">Debug Info</span>
						<button
							type="button"
							onclick={() => (showDebugPanel = false)}
							class="text-gray-400 hover:text-gray-600"
							aria-label="Close debug panel"
						>
							<svg
								class="w-4 h-4"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M6 18L18 6M6 6l12 12"
								></path>
							</svg>
						</button>
					</div>
					<div class="debug-stats">
						<div class="debug-stat">
							<span class="debug-label">Method:</span>
							<span class="debug-value"
								>{scannerReady ? 'BarcodeDetector' : 'Loading...'}</span
							>
						</div>
						<div class="debug-stat">
							<span class="debug-label">Attempts:</span>
							<span class="debug-value">{scanAttempts}</span>
						</div>
						<div class="debug-stat">
							<span class="debug-label">Video:</span>
							<span class="debug-value"
								>{videoElement?.videoWidth || 0}x{videoElement?.videoHeight ||
									0}</span
							>
						</div>
					</div>
					<div class="debug-logs">
						<div class="text-xs font-semibold mb-1">Recent Logs:</div>
						{#each debugLogs as log}
							<div class="debug-log-entry">{log}</div>
						{/each}
						{#if debugLogs.length === 0}
							<div class="debug-log-entry text-gray-400">No logs yet...</div>
						{/if}
					</div>
				</div>
			{:else}
				<button
					type="button"
					onclick={() => (showDebugPanel = true)}
					class="mt-2 text-xs text-gray-400 hover:text-gray-600"
					aria-label="Show debug panel"
				>
					Show Debug
				</button>
			{/if}
		</div>
	</div>
{/if}

<style>
	/* Scanner container - responsive sizing */
	.scanner-container {
		max-width: 28rem; /* 448px - portrait default */
		max-height: 90vh;
		overflow: auto;
	}

	/* Landscape mode - wider container */
	@media (orientation: landscape) {
		.scanner-container {
			max-width: 600px;
			max-height: 80vh;
		}
	}

	/* Video container */
	.scanner-video-container {
		position: relative;
		width: 100%;
		min-height: 500px;
		max-height: 70vh;
		overflow: hidden;
		background-color: #000;
		border-radius: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	@media (orientation: landscape) {
		.scanner-video-container {
			min-height: 400px;
			max-height: 60vh;
		}
	}

	/* Video element */
	video {
		width: 100%;
		height: 100%;
		min-height: 500px;
		object-fit: contain; /* Changed from cover to contain to show full video */
		display: block;
	}

	@media (orientation: landscape) {
		video {
			min-height: 400px;
		}
	}

	/* Scanner overlay */
	.scanner-overlay {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		pointer-events: none;
	}

	/* Scan guide box */
	.scanner-guide {
		width: 70%;
		max-width: 300px;
		height: 100px;
		border: 2px solid rgb(59, 130, 246); /* cyan-500 */
		border-radius: 0.375rem;
		transition: all 0.3s ease;
	}

	/* Searching state - slow pulse */
	.scanner-guide.searching {
		animation: searchingPulse 2s ease-in-out infinite;
	}

	/* Detecting state - faster pulse with orange glow */
	.scanner-guide.detecting {
		border-color: rgb(249, 115, 22); /* orange-500 */
		box-shadow:
			0 0 15px rgba(249, 115, 22, 0.4),
			0 0 30px rgba(249, 115, 22, 0.2);
		animation: detectingPulse 1s ease-in-out infinite;
	}

	/* Success state - green border with glow */
	.scanner-guide.success,
	.scanner-guide.found {
		border-color: rgb(34, 197, 94); /* green-500 */
		box-shadow:
			0 0 20px rgba(34, 197, 94, 0.5),
			0 0 40px rgba(34, 197, 94, 0.3);
		animation: successPulse 0.5s ease-out;
	}

	@keyframes searchingPulse {
		0%,
		100% {
			opacity: 1;
			box-shadow: 0 0 10px rgba(59, 130, 246, 0.3);
		}
		50% {
			opacity: 0.7;
			box-shadow: 0 0 20px rgba(59, 130, 246, 0.5);
		}
	}

	@keyframes detectingPulse {
		0%,
		100% {
			transform: scale(1);
			box-shadow:
				0 0 15px rgba(249, 115, 22, 0.4),
				0 0 30px rgba(249, 115, 22, 0.2);
		}
		50% {
			transform: scale(1.02);
			box-shadow:
				0 0 25px rgba(249, 115, 22, 0.6),
				0 0 50px rgba(249, 115, 22, 0.4);
		}
	}

	@keyframes successPulse {
		0% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.05);
		}
		100% {
			transform: scale(1);
		}
	}

	/* Torch/Flash button */
	.torch-button {
		position: absolute;
		bottom: 1rem;
		right: 1rem;
		padding: 0.75rem;
		background-color: rgba(0, 0, 0, 0.6);
		border-radius: 50%;
		color: white;
		cursor: pointer;
		transition: all 0.2s ease;
		border: none;
		display: flex;
		align-items: center;
		justify-content: center;
		pointer-events: auto;
		z-index: 10;
	}

	.torch-button:hover {
		background-color: rgba(0, 0, 0, 0.8);
		transform: scale(1.1);
	}

	.torch-button.active {
		background-color: rgb(234, 179, 8); /* yellow-600 */
		color: black;
	}

	.torch-button.active:hover {
		background-color: rgb(202, 138, 4); /* yellow-700 */
	}

	/* Loading spinner overlay */
	.scanner-loading {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: rgba(0, 0, 0, 0.5);
		border-radius: 0.5rem;
		z-index: 5;
	}

	/* Scan progress indicator */
	.scan-progress {
		position: absolute;
		bottom: 1rem;
		left: 1rem;
		right: 1rem;
		pointer-events: none;
		z-index: 10;
	}

	.scan-progress-bar {
		width: 100%;
		height: 4px;
		background-color: rgba(255, 255, 255, 0.3);
		border-radius: 2px;
		overflow: hidden;
	}

	.scan-progress-fill {
		height: 100%;
		background: linear-gradient(90deg, rgb(59, 130, 246), rgb(249, 115, 22));
		border-radius: 2px;
		transition: width 0.3s ease;
		animation: progressShimmer 1.5s ease-in-out infinite;
	}

	@keyframes progressShimmer {
		0% {
			opacity: 0.7;
		}
		50% {
			opacity: 1;
		}
		100% {
			opacity: 0.7;
		}
	}

	/* Landscape - smaller guide height */
	@media (orientation: landscape) {
		.scanner-guide {
			height: 80px;
		}
	}

	/* Debug Panel */
	.debug-panel {
		margin-top: 1rem;
		padding: 0.75rem;
		background-color: #1f2937; /* gray-800 */
		color: #f3f4f6; /* gray-100 */
		border-radius: 0.5rem;
		font-family: 'Monaco', 'Courier New', monospace;
		font-size: 0.75rem;
		max-height: 300px;
		overflow-y: auto;
	}

	.debug-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
		padding-bottom: 0.5rem;
		border-bottom: 1px solid #4b5563; /* gray-600 */
	}

	.debug-stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.5rem;
		margin-bottom: 0.75rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid #4b5563; /* gray-600 */
	}

	.debug-stat {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.debug-label {
		color: #9ca3af; /* gray-400 */
		font-size: 0.625rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.debug-value {
		color: #60a5fa; /* cyan-400 */
		font-weight: 600;
	}

	.debug-logs {
		max-height: 150px;
		overflow-y: auto;
	}

	.debug-log-entry {
		padding: 0.25rem;
		margin-bottom: 0.25rem;
		background-color: #374151; /* gray-700 */
		border-radius: 0.25rem;
		word-break: break-all;
		font-size: 0.625rem;
		line-height: 1.2;
	}
</style>
