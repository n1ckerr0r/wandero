import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import { fetchLocations, fetchRoutes } from "./api";
import { SidePanel } from "./SidePanel";
import type {
  Filters,
  GeoJSONFeatureCollection,
  LocationItem,
  RouteItem,
  Tab,
} from "./types";

const emptyGeoJSON: GeoJSONFeatureCollection = {
  type: "FeatureCollection",
  features: [],
};

const MapView = lazy(() =>
  import("./MapView").then((module) => ({ default: module.MapView })),
);

export function App() {
  const [locations, setLocations] = useState<LocationItem[]>([]);
  const [routes, setRoutes] = useState<RouteItem[]>([]);
  const [filters, setFilters] = useState<Filters>({
    q: "",
    from: "",
    to: "",
    tag: "",
  });
  const [activeTab, setActiveTab] = useState<Tab>("locations");
  const [selectedLocationId, setSelectedLocationId] = useState<number>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();
  const [routesError, setRoutesError] = useState<string>();

  const selectedLocation = useMemo(
    () => locations.find((location) => location.id === selectedLocationId),
    [locations, selectedLocationId],
  );

  const visibleLocations = useMemo(() => {
    const search = filters.q.trim().toLocaleLowerCase("ru");
    if (!search) return locations;
    return locations.filter((location) =>
      `${location.name} ${location.description}`
        .toLocaleLowerCase("ru")
        .includes(search),
    );
  }, [filters.q, locations]);

  const reloadLocations = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      setLocations(await fetchLocations());
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Неизвестная ошибка");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void reloadLocations();
  }, [reloadLocations]);

  const reloadRoutes = useCallback(async () => {
    setRoutesError(undefined);
    try {
      setRoutes(await fetchRoutes());
    } catch (caught) {
      setRoutesError(
        caught instanceof Error ? caught.message : "Неизвестная ошибка",
      );
    }
  }, []);

  useEffect(() => {
    void reloadRoutes();
  }, [reloadRoutes]);

  const selectLocation = useCallback((id: number) => {
    setSelectedLocationId(id);
  }, []);

  return (
    <main className="shell">
      <Suspense
        fallback={
          <section className="map-panel map-loading">Загрузка карты…</section>
        }
      >
        <MapView
          events={[]}
          geoJSON={emptyGeoJSON}
          locations={visibleLocations}
          mediaLoading={false}
          selectedEventMedia={[]}
          selectedLocation={selectedLocation}
          onSelectEvent={() => undefined}
          onSelectLocation={selectLocation}
        />
      </Suspense>
      <SidePanel
        activeTab={activeTab}
        error={error}
        filters={filters}
        loading={loading}
        allLocations={locations}
        locations={visibleLocations}
        routes={routes}
        routesError={routesError}
        selectedLocationId={selectedLocationId}
        onChangeFilters={setFilters}
        onRefresh={() => void reloadLocations()}
        onSelectLocation={selectLocation}
        onSetActiveTab={setActiveTab}
        onRouteSaved={() => void reloadRoutes()}
      />
    </main>
  );
}
