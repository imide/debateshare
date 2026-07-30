export class RoomEvents {
	connected = $state(false);

	private es: EventSource | null = null;
	private retryDelay = 1000;
	private retryTimer: number | undefined;
	private refetchTimer: number | undefined;
	private stopped = false;

	constructor(
		private code: string,
		private onChange: () => void,
		private onExpired: () => void
	) {
		this.connect();
	}

	private connect() {
		if (this.stopped) return;
		this.es = new EventSource(`/api/rooms/${this.code}/events`);
		this.es.onopen = () => {
			this.connected = true;
			this.retryDelay = 1000;
			// catchup
			this.scheduleRefetch();
		};
		for (const ev of ['file_added', 'file_updated', 'file_deleted']) {
			this.es.addEventListener(ev, () => this.scheduleRefetch());
		}
		this.es.addEventListener('room_expired', () => this.expire());
		this.es.onerror = () => {
			this.connected = false;
			this.es?.close();
			this.es = null;
			if (this.stopped) return;
			// closed stream means room expired so refresh and get the 404
			this.scheduleRefetch();
			this.retryTimer = window.setTimeout(() => this.connect(), this.retryDelay);
			this.retryDelay = Math.min(this.retryDelay * 2, 30000);
		};
	}

	private scheduleRefetch() {
		clearTimeout(this.refetchTimer);
		this.refetchTimer = window.setTimeout(() => this.onChange(), 150);
	}

	expire() {
		if (this.stopped) return;
		this.close();
		this.onExpired();
	}

	close() {
		this.stopped = true;
		this.connected = false;
		clearTimeout(this.retryTimer);
		clearTimeout(this.refetchTimer);
		this.es?.close();
		this.es = null;
	}
}
