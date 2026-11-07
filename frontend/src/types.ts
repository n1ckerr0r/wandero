import type { components } from "./api-schema";

export type EventItem = components["schemas"]["Event"];
export type EventInput = components["schemas"]["EventInput"];
export type LocationItem = components["schemas"]["Location"];
export type LocationInput = components["schemas"]["LocationInput"];
export type TagItem = components["schemas"]["Tag"];
export type TagInput = components["schemas"]["TagInput"];
export type SourceItem = components["schemas"]["Source"];
export type SourceInput = components["schemas"]["SourceInput"];
export type MediaItem = components["schemas"]["Media"];
export type MediaInput = Omit<components["schemas"]["MediaInput"], "type"> & {
  type?: "image" | "link";
};
export type RouteItem = components["schemas"]["Route"];
export type RouteInput = components["schemas"]["RouteInput"];

export type GeoJSONFeatureCollection = GeoJSON.FeatureCollection<
  GeoJSON.Point,
  Record<string, unknown>
>;

export type Tab = "locations" | "routes";

export type Filters = {
  q: string;
  from: string;
  to: string;
  tag: string;
};
