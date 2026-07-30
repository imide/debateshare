import type { FileInfo, Room } from './types';

export class ApiError extends Error {
	constructor(
		public status: number,
		message: string
	) {
		super(message);
	}
}

// should've probably made a single function with all the shared logic once and then used that function for each query, oh well lol

// default
const json = (body: unknown): RequestInit => ({
	method: 'POST',
	headers: { 'content-type': 'application/json' },
	body: JSON.stringify(body)
});

// Room stuff

export const getRoom = async (code: string): Promise<Room> => {
	const response = await fetch(`/api/rooms/${code}`);
	const data = await response.json();
	if (!response.ok)
		throw new ApiError(response.status, data?.error ?? `request failed (${response.status})`);
	return data as Room;
};

export const createRoom = async (name: string | undefined, email: string): Promise<Room> => {
	const response = await fetch(`/api/rooms`, json({ name, email }));
	const data = await response.json();
	if (!response.ok)
		throw new ApiError(response.status, data?.error ?? `request failed (${response.status})`);
	return data as Room;
};

export const joinRoom = async (code: string, email: string): Promise<Room> => {
	const response = await fetch(`/api/rooms/${code}/join`, json({ email }));
	const data = await response.json();
	if (!response.ok)
		throw new ApiError(response.status, data?.error ?? `request failed (${response.status})`);
	return data as Room;
};

// File stuff

export const listFiles = async (code: string): Promise<FileInfo[]> => {
	const response = await fetch(`/api/rooms/${code}/files`);
	const data = await response.json();
	if (!response.ok)
		throw new ApiError(response.status, data?.error ?? `request failed (${response.status})`);
	return data as FileInfo[];
};

export const deleteFile = async (code: string, id: number): Promise<void> => {
	const response = await fetch(`/api/rooms/${code}/files/${id}`, { method: 'DELETE' });
	const data = await response.json();
	if (!response.ok)
		throw new ApiError(response.status, data?.error ?? `request failed (${response.status})`);
	return data;
};

export const getDownloadUrl = async (code: string, id: number): Promise<string> => {
	const response = await fetch(`/api/rooms/${code}/files/${id}/download`);
	const data = await response.json();
	if (!response.ok)
		throw new ApiError(response.status, data?.error ?? `request failed (${response.status})`);
	return data;
};

export const getRoomZipUrl = async (code: string) => {
	const response = await fetch(`/api/rooms/${code}/files/zip`);
	const data = await response.json();
	if (!response.ok)
		throw new ApiError(response.status, data?.error ?? `request failed (${response.status})`);
	return data;
};

// use xhr to monitor progress.
export const uploadFile = async (
	code: string,
	file: globalThis.File,
	onProgress: (percent: number) => void,
	replaceId?: number
): Promise<FileInfo> => {
	return new Promise((resolve, reject) => {
		const xhr = new XMLHttpRequest();
		const path = replaceId ? `/api/rooms/${code}/files/${replaceId}` : `/api/rooms/${code}/files`;
		xhr.open(replaceId ? 'PUT' : 'POST', path);
		xhr.upload.onprogress = (e) => {
			if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100));
		};
		xhr.onload = () => {
			let body: unknown;
			try {
				body = JSON.parse(xhr.responseText);
			} catch {
				// ignore
			}
			if (xhr.status >= 200 && xhr.status < 300) resolve(body as FileInfo);
			else reject(new ApiError(xhr.status, `upload failed (${xhr.status})`));
		};
		xhr.onerror = () => reject(new ApiError(0, 'network error during upload'));
		const form = new FormData();
		form.append('file', file);
		xhr.send(form);
	});
};
