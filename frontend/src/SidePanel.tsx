import { useMemo, useState } from "react";
import type { Dispatch, SetStateAction } from "react";
import { completeRoute, createRoute } from "./api";
import { Icon } from "./Icon";
import { locationPreviewImage } from "./images";
import type { Filters, LocationItem, RouteItem, Tab } from "./types";

type SidePanelProps = {
  activeTab: Tab;
  error?: string;
  filters: Filters;
  loading: boolean;
  allLocations: LocationItem[];
  locations: LocationItem[];
  routes: RouteItem[];
  routesError?: string;
  selectedLocationId?: number;
  onChangeFilters: Dispatch<SetStateAction<Filters>>;
  onRefresh: () => void;
  onSelectLocation: (id: number) => void;
  onSetActiveTab: (tab: Tab) => void;
  onRouteSaved: () => void;
};

function pluralPlaces(count: number) {
  if (count % 10 === 1 && count % 100 !== 11) return "место";
  if (count % 10 >= 2 && count % 10 <= 4 && !(count % 100 >= 12 && count % 100 <= 14)) return "места";
  return "мест";
}

export function SidePanel({
  activeTab,
  error,
  filters,
  loading,
  allLocations,
  locations,
  routes,
  routesError,
  selectedLocationId,
  onChangeFilters,
  onRefresh,
  onSelectLocation,
  onSetActiveTab,
  onRouteSaved,
}: SidePanelProps) {
  const [selectedRouteId, setSelectedRouteId] = useState<number>();
  const [myWalk, setMyWalk] = useState<number[]>([]);
  const [saveState, setSaveState] = useState<string>();

  const routeLocations = useMemo(
    () => allLocations.filter((location) => myWalk.includes(location.id)),
    [allLocations, myWalk],
  );
  const selectedRoute =
    routes.find((route) => route.id === selectedRouteId) ?? routes[0];

  const toggleWalkPlace = (id: number) => {
    setSaveState(undefined);
    setMyWalk((current) =>
      current.includes(id)
        ? current.filter((item) => item !== id)
        : [...current, id],
    );
  };

  const moveWalkPlace = (id: number, direction: -1 | 1) => {
    setSaveState(undefined);
    setMyWalk((current) => {
      const index = current.indexOf(id);
      const target = index + direction;
      if (index < 0 || target < 0 || target >= current.length) return current;
      const reordered = [...current];
      [reordered[index], reordered[target]] = [
        reordered[target],
        reordered[index],
      ];
      return reordered;
    });
  };

  const saveWalk = async () => {
    if (routeLocations.length === 0) {
      setSaveState("Добавьте хотя бы одно место.");
      return;
    }
    setSaveState("Сохраняем…");
    try {
      await createRoute({
        title: "Моя прогулка",
        description: "Маршрут, собранный в Wandero.",
        published: false,
        stops: routeLocations.map((location) => ({
          location_id: location.id,
          visit_minutes: 30,
        })),
      });
      setSaveState("Прогулка сохранена.");
      onRouteSaved();
    } catch (caught) {
      setSaveState(
        `Не удалось сохранить: ${caught instanceof Error ? caught.message : "неизвестная ошибка"}`,
      );
    }
  };

  return (
    <aside className="side-panel">
      <header className="panel-header">
        <div className="brand-lockup">
          <span className="brand-mark" aria-hidden="true">
            w
          </span>
          <div>
            <p className="eyebrow">Wandero · Петербург</p>
            <h1>Карта города</h1>
          </div>
        </div>
        <button
          className="icon-button refresh-button"
          type="button"
          title="Обновить места"
          aria-label="Обновить места"
          onClick={onRefresh}
        >
          <Icon name="refresh" />
        </button>
      </header>

      <form className="filters" onSubmit={(event) => event.preventDefault()}>
        <label className="search-field">
          <span className="search-glyph" aria-hidden="true">
            <Icon name="search" />
          </span>
          <input
            aria-label="Поиск мест и адресов"
            name="q"
            type="search"
            placeholder="Место или адрес"
            autoComplete="off"
            value={filters.q}
            onChange={(event) =>
              onChangeFilters((current) => ({
                ...current,
                q: event.target.value,
              }))
            }
          />
          <kbd>⌘ K</kbd>
        </label>
      </form>

      <nav className="tabs" aria-label="Разделы карты">
        <button
          className={activeTab === "locations" ? "tab active" : "tab"}
          type="button"
          onClick={() => onSetActiveTab("locations")}
        >
          <Icon name="map" /> Места
        </button>
        <button
          className={activeTab === "routes" ? "tab active" : "tab"}
          type="button"
          onClick={() => onSetActiveTab("routes")}
        >
          <Icon name="route" /> Маршруты
        </button>
      </nav>

      <div className="panel-content">
        {activeTab === "locations" ? (
          <>
            <div className="section-heading">
              <div>
                <p className="eyebrow">Исследуйте рядом</p>
                <h2>Места в Петербурге</h2>
              </div>
              <span className="count-pill">{locations.length}</span>
            </div>
            <p className="section-note">
              {error
                ? `Не удалось загрузить места: ${error}`
                : loading
                  ? "Обновляем карту…"
                  : "Выберите точку на карте или откройте место из списка."}
            </p>
            <LocationList
              locations={locations}
              selectedLocationId={selectedLocationId}
              onSelectLocation={onSelectLocation}
              onAddToWalk={toggleWalkPlace}
              addedIds={myWalk}
            />
          </>
        ) : (
          <>
            <div className="section-heading">
              <div>
                <p className="eyebrow">Проверено местными</p>
                <h2>Популярные прогулки</h2>
              </div>
              <span className="count-pill">{routes.length}</span>
            </div>

            <div className="route-list">
              {routes.map((route, index) => (
                <button
                  className={
                    selectedRoute?.id === route.id
                      ? "route-card selected"
                      : "route-card"
                  }
                  key={route.id}
                  type="button"
                  onClick={() => setSelectedRouteId(route.id)}
                >
                  <span
                    className={`route-art route-art-${index + 1}`}
                    aria-hidden="true"
                  >
                    <span className="route-art-line" />
                    <span className="route-art-pin route-art-pin-first" />
                    <span className="route-art-pin route-art-pin-last" />
                  </span>
                  <span className="route-copy">
                    <span className="route-category">Редакционный маршрут</span>
                    <strong>{route.title}</strong>
                    <span className="route-meta">
                      {route.walking_minutes != null
                        ? `${route.walking_minutes} мин пешком`
                        : "Время не указано"}
                      {route.distance_meters != null ? <i /> : null}
                      {route.distance_meters != null
                        ? `${(route.distance_meters / 1000).toLocaleString("ru-RU", { maximumFractionDigits: 1 })} км`
                        : null}
                      <i />
                      {route.stops.length} {pluralPlaces(route.stops.length)}
                    </span>
                    <span className="walked-count">
                      <Icon name="walk" />
                      {route.completion_count} прошли
                    </span>
                  </span>
                  <span className="route-arrow" aria-hidden="true">
                    <Icon name="route" />
                  </span>
                </button>
              ))}
            </div>
            {routesError ? (
              <p className="section-note">
                Не удалось загрузить маршруты: {routesError}
              </p>
            ) : null}
            {!routesError && routes.length === 0 ? (
              <p className="empty">Готовых маршрутов пока нет.</p>
            ) : null}

            {selectedRoute ? (
              <section className="route-detail" aria-live="polite">
                <div className="detail-heading">
                  <div>
                    <p className="eyebrow">Маршрут дня</p>
                    <h3>{selectedRoute.title}</h3>
                  </div>
                  <span className="route-detail-icon" aria-hidden="true">
                    <Icon name="route" />
                  </span>
                </div>
                <p>{selectedRoute.description}</p>
                <ol className="itinerary">
                  {selectedRoute.stops.map((stop) => (
                    <li key={stop.position}>{stop.location.name}</li>
                  ))}
                </ol>
                <button
                  className="add-place"
                  type="button"
                  onClick={() => {
                    void (async () => {
                      try {
                        await completeRoute(selectedRoute.id);
                        onRouteSaved();
                      } catch (caught) {
                        setSaveState(
                          `Не удалось отметить прогулку: ${caught instanceof Error ? caught.message : "неизвестная ошибка"}`,
                        );
                      }
                    })();
                  }}
                >
                  Я прошёл маршрут
                </button>
              </section>
            ) : null}

            <section className="my-walk">
              <div className="detail-heading">
                <div>
                  <p className="eyebrow">Соберите свою прогулку</p>
                  <h3>
                    Мой маршрут <span>{routeLocations.length}</span>
                  </h3>
                </div>
                <span
                  className="route-detail-icon custom-icon"
                  aria-hidden="true"
                >
                  <Icon name="add" />
                </span>
              </div>
              {routeLocations.length > 0 ? (
                <ol className="itinerary personal-itinerary">
                  {routeLocations.map((location, index) => (
                    <li key={location.id}>
                      <span>{index + 1}</span>
                      <button
                        type="button"
                        onClick={() => toggleWalkPlace(location.id)}
                        aria-label={`Убрать ${location.name}`}
                      >
                        {location.name}
                        <Icon name="close" />
                      </button>
                      <button
                        type="button"
                        disabled={index === 0}
                        onClick={() => moveWalkPlace(location.id, -1)}
                        aria-label={`Поднять ${location.name}`}
                      >
                        <Icon name="arrowUp" />
                      </button>
                      <button
                        type="button"
                        disabled={index === routeLocations.length - 1}
                        onClick={() => moveWalkPlace(location.id, 1)}
                        aria-label={`Опустить ${location.name}`}
                      >
                        <Icon name="arrowDown" />
                      </button>
                    </li>
                  ))}
                </ol>
              ) : (
                <p>Добавляйте интересные места из списка ниже.</p>
              )}
              <div className="route-place-picker">
                {locations.slice(0, 6).map((location) => (
                  <button
                    className={
                      myWalk.includes(location.id)
                        ? "place-option added"
                        : "place-option"
                    }
                    key={location.id}
                    type="button"
                    onClick={() => toggleWalkPlace(location.id)}
                  >
                    <span>{location.name}</span>
                    <b>{myWalk.includes(location.id) ? "✓" : "+"}</b>
                  </button>
                ))}
                {locations.length === 0 ? (
                  <span className="empty">Места пока не загрузились</span>
                ) : null}
              </div>
              <button
                className="add-place"
                type="button"
                onClick={() => void saveWalk()}
                disabled={routeLocations.length === 0}
              >
                Сохранить прогулку
              </button>
              {saveState ? (
                <p className="section-note" aria-live="polite">
                  {saveState}
                </p>
              ) : null}
            </section>
          </>
        )}
      </div>
    </aside>
  );
}

function LocationList({
  addedIds,
  locations,
  onAddToWalk,
  onSelectLocation,
  selectedLocationId,
}: {
  addedIds: number[];
  locations: LocationItem[];
  onAddToWalk: (id: number) => void;
  onSelectLocation: (id: number) => void;
  selectedLocationId?: number;
}) {
  if (locations.length === 0) {
    return (
      <div className="empty">Места не найдены. Попробуйте другой запрос.</div>
    );
  }

  return (
    <div className="location-list">
      {locations.map((location) => {
        const imageUrl = locationPreviewImage(location);
        const added = addedIds.includes(location.id);
        return (
          <article
            key={location.id}
            className={
              location.id === selectedLocationId
                ? "location-card active"
                : "location-card"
            }
          >
            <button
              className="location-main"
              type="button"
              onClick={() => onSelectLocation(location.id)}
            >
              {imageUrl ? (
                <img src={imageUrl} alt="" loading="lazy" />
              ) : (
                <span className="location-image-placeholder" aria-hidden="true">
                  <Icon name="map" />
                </span>
              )}
              <span className="location-copy">
                <strong>{location.name}</strong>
                <span>
                  {location.description ||
                    "Историческое место в центре Петербурга."}
                </span>
              </span>
            </button>
            <button
              className={added ? "add-place added" : "add-place"}
              type="button"
              onClick={() => onAddToWalk(location.id)}
              aria-label={
                added
                  ? `Убрать ${location.name} из маршрута`
                  : `Добавить ${location.name} в маршрут`
              }
              title={added ? "Убрать из маршрута" : "Добавить в мою прогулку"}
            >
              <Icon name={added ? "check" : "add"} />
            </button>
          </article>
        );
      })}
    </div>
  );
}
