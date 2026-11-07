import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createEvent,
  createEventMedia,
  createLocation,
  createRoute,
  createSource,
  createTag,
  deleteEvent,
  deleteLocation,
  deleteRoute,
  deleteMedia,
  deleteSource,
  deleteTag,
  fetchEvent,
  fetchEventMedia,
  fetchEvents,
  fetchEventsGeoJSON,
  fetchEventSources,
  fetchHealth,
  fetchLocation,
  fetchLocations,
  fetchRoute,
  fetchRoutes,
  fetchSource,
  fetchSources,
  fetchTags,
  linkEventSource,
  setAPIWriteToken,
  unlinkEventSource,
  updateEvent,
  updateLocation,
  updateRoute,
  completeRoute,
  updateSource,
  updateTag,
} from "./api";
import type { paths } from "./api-schema";

const locationInput = {
  name: "Казанский собор",
  description: "Исторический памятник",
  latitude: 59.9343,
  longitude: 30.3241,
};
const eventInput = {
  title: "Открытие",
  date_from: "1801-09-01",
  tags: ["история"],
};
const sourceInput = {
  title: "Архив",
  author: "ЦГИА",
  type: "archive" as const,
};
const tagInput = { name: "архитектура" };
const routeInput = {
  title: "Моя прогулка",
  published: false,
  stops: [{ location_id: 1, visit_minutes: 30 }],
};

const backendOperations = {
  "/health": ["get"],
  "/api/locations": ["get", "post"],
  "/api/locations/{id}": ["get", "put", "delete"],
  "/api/routes": ["get", "post"],
  "/api/routes/{id}": ["get", "put", "delete"],
  "/api/routes/{id}/completions": ["post"],
  "/api/events": ["get", "post"],
  "/api/events/{id}": ["get", "put", "delete"],
  "/api/events/{id}/media": ["get", "post"],
  "/api/events/{id}/sources": ["get"],
  "/api/events/{id}/sources/{source_id}": ["post", "delete"],
  "/api/tags": ["get", "post"],
  "/api/tags/{id}": ["put", "delete"],
  "/api/sources": ["get", "post"],
  "/api/sources/{id}": ["get", "put", "delete"],
  "/api/media/{id}": ["delete"],
  "/api/map/events.geojson": ["get"],
} as const satisfies {
  [Path in keyof paths]: readonly DeclaredMethods<Path>[];
};

type DeclaredMethods<Path extends keyof paths> = {
  [Method in Exclude<keyof paths[Path], "parameters">]: NonNullable<
    paths[Path][Method]
  > extends never
    ? never
    : Method;
}[Exclude<keyof paths[Path], "parameters">];

type MissingOpenAPIMethods = {
  [Path in keyof paths]: Exclude<
    DeclaredMethods<Path>,
    (typeof backendOperations)[Path][number]
  >;
}[keyof paths];
const openAPICoverageIsComplete: [MissingOpenAPIMethods] extends [never]
  ? true
  : never = true;
export const generatedOpenAPICoverage = openAPICoverageIsComplete;

const operations = [
  ["health", () => fetchHealth(), "GET", "/health"],
  ["list tags", () => fetchTags(), "GET", "/api/tags"],
  ["create tag", () => createTag(tagInput), "POST", "/api/tags"],
  ["update tag", () => updateTag(4, tagInput), "PUT", "/api/tags/4"],
  ["delete tag", () => deleteTag(4), "DELETE", "/api/tags/4"],
  ["list locations", () => fetchLocations(), "GET", "/api/locations"],
  [
    "create location",
    () => createLocation(locationInput),
    "POST",
    "/api/locations",
  ],
  ["get location", () => fetchLocation(8), "GET", "/api/locations/8"],
  [
    "update location",
    () => updateLocation(8, locationInput),
    "PUT",
    "/api/locations/8",
  ],
  ["delete location", () => deleteLocation(8), "DELETE", "/api/locations/8"],
  ["list routes", () => fetchRoutes(), "GET", "/api/routes"],
  ["create route", () => createRoute(routeInput), "POST", "/api/routes"],
  ["get route", () => fetchRoute(8), "GET", "/api/routes/8"],
  ["update route", () => updateRoute(8, routeInput), "PUT", "/api/routes/8"],
  ["delete route", () => deleteRoute(8), "DELETE", "/api/routes/8"],
  [
    "complete route",
    () => completeRoute(8),
    "POST",
    "/api/routes/8/completions",
  ],
  ["list events", () => fetchEvents(), "GET", "/api/events"],
  ["create event", () => createEvent(eventInput), "POST", "/api/events"],
  ["get event", () => fetchEvent(12), "GET", "/api/events/12"],
  ["update event", () => updateEvent(12, eventInput), "PUT", "/api/events/12"],
  ["delete event", () => deleteEvent(12), "DELETE", "/api/events/12"],
  [
    "list event media",
    () => fetchEventMedia(12),
    "GET",
    "/api/events/12/media",
  ],
  [
    "create event media",
    () => createEventMedia(12, { url: "https://example.test/image.jpg" }),
    "POST",
    "/api/events/12/media",
  ],
  ["delete media", () => deleteMedia(13), "DELETE", "/api/media/13"],
  [
    "list event sources",
    () => fetchEventSources(12),
    "GET",
    "/api/events/12/sources",
  ],
  [
    "link event source",
    () => linkEventSource(12, 15),
    "POST",
    "/api/events/12/sources/15",
  ],
  [
    "unlink event source",
    () => unlinkEventSource(12, 15),
    "DELETE",
    "/api/events/12/sources/15",
  ],
  ["list sources", () => fetchSources(), "GET", "/api/sources"],
  ["create source", () => createSource(sourceInput), "POST", "/api/sources"],
  ["get source", () => fetchSource(15), "GET", "/api/sources/15"],
  [
    "update source",
    () => updateSource(15, sourceInput),
    "PUT",
    "/api/sources/15",
  ],
  ["delete source", () => deleteSource(15), "DELETE", "/api/sources/15"],
  [
    "events GeoJSON",
    () => fetchEventsGeoJSON(),
    "GET",
    "/api/map/events.geojson",
  ],
] as const;

describe("generated-contract API client", () => {
  const fetchMock = vi.fn<typeof fetch>();

  beforeEach(() => {
    fetchMock.mockReset();
    setAPIWriteToken(undefined);
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("has a client test for every generated OpenAPI method", () => {
    const generatedMethodCount = Object.values(backendOperations).reduce(
      (total, methods) => total + methods.length,
      0,
    );
    expect(operations).toHaveLength(generatedMethodCount);
  });

  it.each(operations)(
    "maps %s to its backend operation",
    async (name, call, method, path) => {
      fetchMock.mockResolvedValue(
        method === "DELETE" ||
          name === "link event source" ||
          name === "unlink event source"
          ? new Response(null, { status: 204 })
          : Response.json([]),
      );

      await call();

      expect(fetchMock).toHaveBeenCalledTimes(1);
      const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
      expect(url).toBe(path);
      expect(init.method ?? "GET").toBe(method);
      expect(new Headers(init.headers).get("Accept")).toBe("application/json");
      if (
        (method === "POST" &&
          name !== "link event source" &&
          name !== "complete route") ||
        method === "PUT"
      ) {
        if (typeof init.body !== "string") {
          throw new Error("Expected a JSON request body");
        }
        expect(JSON.parse(init.body)).toBeDefined();
        expect(new Headers(init.headers).get("Content-Type")).toBe(
          "application/json",
        );
      }
    },
  );

  it("preserves event filters and URL-encodes query values", async () => {
    fetchMock.mockResolvedValue(Response.json([]));
    await fetchEvents(
      new URLSearchParams({ q: "Казанский собор", from: "1700" }),
    );
    expect(fetchMock.mock.calls[0]?.[0]).toBe(
      "/api/events?q=%D0%9A%D0%B0%D0%B7%D0%B0%D0%BD%D1%81%D0%BA%D0%B8%D0%B9+%D1%81%D0%BE%D0%B1%D0%BE%D1%80&from=1700",
    );
  });

  it("adds the in-memory bearer token only to write operations", async () => {
    fetchMock.mockImplementation(() =>
      Promise.resolve(Response.json({ id: 1, name: "tag" }, { status: 201 })),
    );
    setAPIWriteToken("  short-lived-token  ");
    await fetchLocations();
    await createTag(tagInput);

    expect(
      new Headers(fetchMock.mock.calls[0]?.[1]?.headers).has("Authorization"),
    ).toBe(false);
    expect(
      new Headers(fetchMock.mock.calls[1]?.[1]?.headers).get("Authorization"),
    ).toBe("Bearer short-lived-token");
  });

  it("surfaces API JSON errors and gives a useful fallback for non-JSON errors", async () => {
    fetchMock.mockResolvedValueOnce(
      Response.json({ error: "место не найдено" }, { status: 404 }),
    );
    await expect(fetchLocation(9)).rejects.toThrow("место не найдено");

    fetchMock.mockResolvedValueOnce(
      new Response("proxy failure", { status: 502 }),
    );
    await expect(fetchLocation(9)).rejects.toThrow("API вернул ошибку 502");
  });

  it("rejects invalid path IDs before making a request", () => {
    expect(() => fetchLocation(0)).toThrow(RangeError);
    expect(() => deleteTag(Number.MAX_SAFE_INTEGER + 1)).toThrow(RangeError);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
