import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SidePanel } from "./SidePanel";
import type { Filters, LocationItem, RouteItem, Tab } from "./types";

const places: LocationItem[] = [
  {
    id: 1,
    name: "Казанский собор",
    description: "Собор на Невском проспекте.",
    latitude: 59.9343,
    longitude: 30.3241,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    name: "Дворцовая площадь",
    description: "Главная площадь Петербурга.",
    latitude: 59.9398,
    longitude: 30.3146,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const routes: RouteItem[] = [
  {
    id: 10,
    title: "Парадный Петербург",
    description: "Прогулка по центру.",
    published: true,
    completion_count: 12,
    stops: [
      { position: 1, location_id: 1, visit_minutes: 30, location: places[0] },
      { position: 2, location_id: 2, visit_minutes: 30, location: places[1] },
    ],
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

afterEach(() => cleanup());

function setup(
  overrides: Partial<{
    activeTab: Tab;
    error: string;
    filters: Filters;
    loading: boolean;
    locations: LocationItem[];
    selectedLocationId: number;
  }> = {},
) {
  const props = {
    activeTab: "locations" as Tab,
    filters: { q: "", from: "", to: "", tag: "" },
    loading: false,
    allLocations: places,
    locations: places,
    routes,
    onChangeFilters: vi.fn(),
    onRefresh: vi.fn(),
    onSelectLocation: vi.fn(),
    onSetActiveTab: vi.fn(),
    onRouteSaved: vi.fn(),
    ...overrides,
  };
  const result = render(<SidePanel {...props} />);
  return { ...result, props };
}

describe("SidePanel", () => {
  it("shows place search and no events tab in the locations view", () => {
    setup();

    expect(
      screen.getByRole("searchbox", { name: "Поиск мест и адресов" }),
    ).toBeTruthy();
    expect(
      screen.getByRole("heading", { name: "Места в Петербурге" }),
    ).toBeTruthy();
    expect(screen.queryByRole("button", { name: /События/ })).toBeNull();
    expect(screen.getByText("Казанский собор")).toBeTruthy();
  });

  it("reports place selection and adds or removes a location from the custom walk", async () => {
    const user = userEvent.setup();
    const { props } = setup();

    await user.click(
      screen.getByRole("button", { name: /Казанский соборСобор/ }),
    );
    expect(props.onSelectLocation).toHaveBeenCalledWith(1);

    await user.click(
      screen.getByRole("button", {
        name: "Добавить Казанский собор в маршрут",
      }),
    );
    expect(
      screen.getByRole("button", {
        name: "Убрать Казанский собор из маршрута",
      }),
    ).toBeTruthy();
    await user.click(
      screen.getByRole("button", {
        name: "Убрать Казанский собор из маршрута",
      }),
    );
    expect(
      screen.getByRole("button", {
        name: "Добавить Казанский собор в маршрут",
      }),
    ).toBeTruthy();
  });

  it("switches to popular routes and changes the selected itinerary", async () => {
    const user = userEvent.setup();
    const { rerender, props } = setup();

    await user.click(screen.getByRole("button", { name: /Маршруты/ }));
    rerender(<SidePanel {...props} activeTab="routes" />);
    expect(
      screen.getByRole("heading", { name: "Популярные прогулки" }),
    ).toBeTruthy();
    expect(screen.getByText(/12 прошли/)).toBeTruthy();

    expect(
      screen.getByRole("heading", { name: "Парадный Петербург" }),
    ).toBeTruthy();
    const itinerary = screen.getByRole("list");
    expect(within(itinerary).getByText("Казанский собор")).toBeTruthy();
  });

  it("keeps selected places visible when the search result set changes", async () => {
    const user = userEvent.setup();
    const { rerender, props } = setup();

    await user.click(screen.getByRole("button", { name: /Маршруты/ }));
    rerender(<SidePanel {...props} activeTab="routes" />);
    await user.click(
      screen.getByRole("button", { name: /^Казанский собор\+/ }),
    );
    rerender(
      <SidePanel
        {...props}
        activeTab="routes"
        locations={[places[1]]}
        filters={{ ...props.filters, q: "Дворцовая" }}
      />,
    );

    const myWalk = screen
      .getByRole("heading", { name: /Мой маршрут/ })
      .closest("section");
    expect(myWalk).toBeTruthy();
    expect(
      within(myWalk as HTMLElement).getByText("Казанский собор"),
    ).toBeTruthy();
    expect(
      within(myWalk as HTMLElement).getByText("Дворцовая площадь"),
    ).toBeTruthy();
  });

  it("shows loading, fetch errors, and empty search results", () => {
    const { rerender, props } = setup({ loading: true });
    expect(screen.getByText("Обновляем карту…")).toBeTruthy();

    rerender(<SidePanel {...props} error="нет связи" locations={[]} />);
    expect(
      screen.getByText("Не удалось загрузить места: нет связи"),
    ).toBeTruthy();
    expect(
      screen.getByText("Места не найдены. Попробуйте другой запрос."),
    ).toBeTruthy();
  });
});
