import type { FullConfig } from '@playwright/test';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

const HEALTH_URL = process.env.BASE_URL
	? `${process.env.BASE_URL.replace(/\/$/, '')}/health`
	: 'https://e2e.savvy.test/health';

/**
 * Skip the verification when Playwright runs in info-only mode or when the
 * caller explicitly opts out. Keeps `--list`, `--help`, `--version` instant
 * (no curl), and lets CI/dev override with SKIP_E2E_SETUP=true if the stack
 * is known-good or being verified separately.
 */
function shouldSkipSetup(): boolean {
	const isListMode = process.argv.includes('--list');
	const isHelpMode =
		process.argv.includes('--help') || process.argv.includes('-h');
	const isVersionMode =
		process.argv.includes('--version') || process.argv.includes('-v');
	const skipSetup = process.env.SKIP_E2E_SETUP === 'true';

	return isListMode || isHelpMode || isVersionMode || skipSetup;
}

/**
 * Probe the e2e app's HTTP health endpoint via curl. `-k` tolerates the
 * dde-mkcert TLS cert; `-fS` makes curl fail loudly on >=400 status and
 * stay quiet on success.
 *
 * Retries a few times because the dde traefik occasionally needs a moment
 * to register a freshly attached container's route after `e2e:start`,
 * even when the container is already health=healthy on the docker side.
 */
async function probeHealthEndpoint(): Promise<void> {
	const maxAttempts = 5;
	const delayMs = 1000;
	let lastError: unknown;
	for (let attempt = 1; attempt <= maxAttempts; attempt++) {
		try {
			await execAsync(`curl -ksSf -o /dev/null --max-time 5 "${HEALTH_URL}"`);
			return;
		} catch (error) {
			lastError = error;
			if (attempt < maxAttempts) {
				await new Promise((resolve) => setTimeout(resolve, delayMs));
			}
		}
	}
	throw lastError instanceof Error ? lastError : new Error(String(lastError));
}

/**
 * Global setup — runs once before all tests.
 *
 * The E2E stack is provisioned externally (`dde project:e2e:start`),
 * not from inside the test runner. This hook only verifies that the stack
 * is reachable and fails fast with a copy-pasteable hint when it isn't.
 */
async function globalSetup(_config: FullConfig): Promise<void> {
	if (shouldSkipSetup()) {
		console.log(
			'ℹ️  Skipping global setup (info-only command or SKIP_E2E_SETUP=true)'
		);
		return;
	}

	try {
		await probeHealthEndpoint();
		console.log(`✓ E2E stack reachable at ${HEALTH_URL}`);
	} catch (error) {
		const message = error instanceof Error ? error.message : String(error);
		throw new Error(
			`E2E stack is not reachable at ${HEALTH_URL}.\n\n` +
				`Most likely fix:\n` +
				`  dde project:e2e:start\n\n` +
				`If the stack is up but traefik is not routing (rare, after long\n` +
				`local sessions): docker restart dde-traefik\n\n` +
				`Underlying error: ${message}`,
			{ cause: error }
		);
	}
}

export default globalSetup;
