import type {
  EventInput,
  EventItem,
  GeoJSONFeatureCollection,
  LocationInput,
  LocationItem,
  MediaInput,
  MediaItem,
  RouteInput,
  RouteItem,
  SourceInput,
  SourceItem,
  TagInput,
  TagItem,
} from "./types";
import type { paths } from "./api-schema";

type ApiPath = keyof paths;

const configuredBaseURL: unknown = import.meta.env.VITE_API_BASE_URL;
const API_BASE_URL = (
  typeof configuredBaseURL === "string" ? configuredBaseURL : ""
).replace(/\/$/, "");
let writeToken: string | undefined;

// Store short-lived credentials in memory; never bake the server write key into Vite env or a bundle.
export function setAPIWriteToken(token: string | undefined): void {
  writeToken = token?.trim() || undefined;
}

export type HealthStatus = { status: "ok" };

export function fetchHealth(): Promise<HealthStatus> {
  return fetchJSON(apiPath("/health"));
}

export function fetchTags(): Promise<TagItem[]> {
  return fetchJSON(apiPath("/api/tags"));
}

export function createTag(input: TagInput): Promise<TagItem> {
  return sendJSON("POST", apiPath("/api/tags"), input);
}

export function updateTag(id: number, input: TagInput): Promise<TagItem> {
  return sendJSON(
    "PUT",
    apiPath("/api/tags/{id}", { id: positiveId(id) }),
    input,
  );
}

export function deleteTag(id: number): Promise<void> {
  return sendVoid("DELETE", apiPath("/api/tags/{id}", { id: positiveId(id) }));
}

export function fetchLocations(): Promise<LocationItem[]> {
  return fetchJSON(apiPath("/api/locations"));
}

export function createLocation(input: LocationInput): Promise<LocationItem> {
  return sendJSON("POST", apiPath("/api/locations"), input);
}

export function fetchLocation(id: number): Promise<LocationItem> {
  return fetchJSON(apiPath("/api/locations/{id}", { id: positiveId(id) }));
}

export function updateLocation(
  id: number,
  input: LocationInput,
): Promise<LocationItem> {
  return sendJSON(
    "PUT",
    apiPath("/api/locations/{id}", { id: positiveId(id) }),
    input,
  );
}

export function deleteLocation(id: number): Promise<void> {
  return sendVoid(
    "DELETE",
    apiPath("/api/locations/{id}", { id: positiveId(id) }),
  );
}

export function fetchRoutes(query = ""): Promise<RouteItem[]> {
  const params = new URLSearchParams();
  if (query.trim()) params.set("q", query.trim());
  return fetchJSON(withQuery(apiPath("/api/routes"), params));
}

export function fetchRoute(id: number): Promise<RouteItem> {
  return fetchJSON(apiPath("/api/routes/{id}", { id: positiveId(id) }));
}

export function createRoute(input: RouteInput): Promise<RouteItem> {
  return sendJSON("POST", apiPath("/api/routes"), input);
}

export function updateRoute(id: number, input: RouteInput): Promise<RouteItem> {
  return sendJSON(
    "PUT",
    apiPath("/api/routes/{id}", { id: positiveId(id) }),
    input,
  );
}

export function deleteRoute(id: number): Promise<void> {
  return sendVoid(
    "DELETE",
    apiPath("/api/routes/{id}", { id: positiveId(id) }),
  );
}

export function completeRoute(id: number): Promise<RouteItem> {
  return sendJSON(
    "POST",
    apiPath("/api/routes/{id}/completions", { id: positiveId(id) }),
    undefined,
  );
}

export function fetchEvents(
  query: URLSearchParams = new URLSearchParams(),
): Promise<EventItem[]> {
  return fetchJSON(withQuery(apiPath("/api/events"), query));
}

export function createEvent(input: EventInput): Promise<EventItem> {
  return sendJSON("POST", apiPath("/api/events"), input);
}

export function fetchEvent(id: number): Promise<EventItem> {
  return fetchJSON(apiPath("/api/events/{id}", { id: positiveId(id) }));
}

export function updateEvent(id: number, input: EventInput): Promise<EventItem> {
  return sendJSON(
    "PUT",
    apiPath("/api/events/{id}", { id: positiveId(id) }),
    input,
  );
}

export function deleteEvent(id: number): Promise<void> {
  return sendVoid(
    "DELETE",
    apiPath("/api/events/{id}", { id: positiveId(id) }),
  );
}

export function fetchEventMedia(eventId: number): Promise<MediaItem[]> {
  return fetchJSON(
    apiPath("/api/events/{id}/media", { id: positiveId(eventId) }),
  );
}

export function createEventMedia(
  eventId: number,
  input: MediaInput,
): Promise<MediaItem> {
  return sendJSON(
    "POST",
    apiPath("/api/events/{id}/media", { id: positiveId(eventId) }),
    input,
  );
}

export function deleteMedia(id: number): Promise<void> {
  return sendVoid("DELETE", apiPath("/api/media/{id}", { id: positiveId(id) }));
}

export function fetchEventSources(eventId: number): Promise<SourceItem[]> {
  return fetchJSON(
    apiPath("/api/events/{id}/sources", { id: positiveId(eventId) }),
  );
}

export function linkEventSource(
  eventId: number,
  sourceId: number,
): Promise<void> {
  return sendVoid(
    "POST",
    apiPath("/api/events/{id}/sources/{source_id}", {
      id: positiveId(eventId),
      source_id: positiveId(sourceId),
    }),
  );
}

export function unlinkEventSource(
  eventId: number,
  sourceId: number,
): Promise<void> {
  return sendVoid(
    "DELETE",
    apiPath("/api/events/{id}/sources/{source_id}", {
      id: positiveId(eventId),
      source_id: positiveId(sourceId),
    }),
  );
}

export function fetchSources(): Promise<SourceItem[]> {
  return fetchJSON(apiPath("/api/sources"));
}

export function createSource(input: SourceInput): Promise<SourceItem> {
  return sendJSON("POST", apiPath("/api/sources"), input);
}

export function fetchSource(id: number): Promise<SourceItem> {
  return fetchJSON(apiPath("/api/sources/{id}", { id: positiveId(id) }));
}

export function updateSource(
  id: number,
  input: SourceInput,
): Promise<SourceItem> {
  return sendJSON(
    "PUT",
    apiPath("/api/sources/{id}", { id: positiveId(id) }),
    input,
  );
}

export function deleteSource(id: number): Promise<void> {
  return sendVoid(
    "DELETE",
    apiPath("/api/sources/{id}", { id: positiveId(id) }),
  );
}

export function fetchEventsGeoJSON(
  query: URLSearchParams = new URLSearchParams(),
): Promise<GeoJSONFeatureCollection> {
  return fetchJSON(withQuery(apiPath("/api/map/events.geojson"), query));
}

async function fetchJSON<T>(path: string): Promise<T> {
  const response = await request(path);
  const body: unknown = await response.json().catch(() => undefined);
  if (!response.ok) {
    throw responseError(response.status, body);
  }
  if (body === undefined) {
    throw new Error("API вернул пустой JSON-ответ");
  }
  return body as T;
}

async function sendJSON<T>(
  method: "POST" | "PUT",
  path: string,
  payload: unknown,
): Promise<T> {
  const response = await request(path, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const body: unknown = await response.json().catch(() => undefined);
  if (!response.ok) {
    throw responseError(response.status, body);
  }
  if (body === undefined) {
    throw new Error("API вернул пустой JSON-ответ");
  }
  return body as T;
}

async function sendVoid(
  method: "POST" | "DELETE",
  path: string,
): Promise<void> {
  const response = await request(path, { method });
  if (!response.ok) {
    const body: unknown = await response.json().catch(() => undefined);
    throw responseError(response.status, body);
  }
}

function request(path: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (writeToken && init.method && init.method !== "GET") {
    headers.set("Authorization", `Bearer ${writeToken}`);
  }
  return fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
    signal: AbortSignal.timeout(15_000),
  });
}

function withQuery(path: string, query: URLSearchParams): string {
  const value = query.toString();
  return value ? `${path}?${value}` : path;
}

function apiPath<P extends ApiPath>(
  template: P,
  params: Record<string, string | number> = {},
): string {
  return template.replace(/\{([^}]+)\}/g, (_match, key: string) => {
    const value = params[key];
    if (value === undefined) {
      throw new Error(`Missing path parameter: ${key}`);
    }
    return encodeURIComponent(String(value));
  });
}

function positiveId(id: number): number {
  if (!Number.isSafeInteger(id) || id <= 0) {
    throw new RangeError("id must be a positive safe integer");
  }
  return id;
}

function responseError(status: number, body: unknown): Error {
  if (
    typeof body === "object" &&
    body !== null &&
    "error" in body &&
    typeof body.error === "string"
  ) {
    return new Error(body.error);
  }
  return new Error(`API вернул ошибку ${status}`);
}
