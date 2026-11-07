# Wandero — Figma handoff

Этот документ — точная спецификация для переноса текущего frontend в редактируемый Figma-файл и развития продукта в приложение для городских прогулок.

## Структура Figma-файла

Создать страницы:

1. `00 Cover & flows`
2. `01 Foundations`
3. `02 Components`
4. `03 Current product`
5. `04 Walk planner — desktop`
6. `05 Walk planner — mobile`
7. `06 Popular routes`
8. `07 Prototype`

Все экраны строить на Auto Layout. Desktop grid: 12 колонок, margin 32, gutter 24. Mobile grid: 4 колонки, margin 16, gutter 12. Базовый шаг отступов — 4 px.

## Foundations из текущего frontend

### Цвета

| Token | Value | Использование |
| --- | --- | --- |
| `Surface/App` | `#EDE8DF` | фон карты/оболочки |
| `Surface/Panel` | `#F9F7F2` | боковая панель |
| `Surface/Card` | `#FFFFFF` | карточки и popup |
| `Surface/Muted` | `#F3F7F1` | исторический контекст |
| `Text/Primary` | `#17201A` | заголовки |
| `Text/Secondary` | `#404942` | основной текст |
| `Text/Muted` | `#6E756F` | подписи |
| `Brand/Primary` | `#2F6B54` | действия, ссылки, маршрут |
| `Brand/Historical` | `#7A3F20` | исторический акцент |
| `Map/Event` | `#C2410C` | событие на карте |
| `Map/Selected` | `#FACC15` | выбранная точка |
| `Border/Default` | `#D8D0C4` | границы |
| `State/Error` | `#B42318` | ошибки |

### Типографика

Inter / system sans. Стили: `Display/28/34 Semibold`, `H1/24/27 Bold`, `H2/18/23 Bold`, `Body/14/20 Regular`, `Body Strong/14/20 Semibold`, `Caption/12/16`, `Button/14/18 Semibold`.

### Геометрия

- Основной radius: 8 px; route cards: 12 px; mobile sheets: 20 px сверху.
- Тень карточки: `0 12 28 / 10%`, цвет `#1A1611`.
- Минимальная интерактивная область: 44×44 px.
- Desktop side panel: 440 px. Карта занимает остальную ширину.

## Components

Создать variants и properties:

- `Button`: primary / secondary / ghost / icon; default / hover / pressed / disabled / loading.
- `Segmented control`: События / Места / Маршруты.
- `Search field`: empty / value / focused / error.
- `Filter chip`: category, duration, indoor, free, family; selected/unselected.
- `Place card`: compact / image; default / selected / added; кнопка `+ В маршрут`.
- `Event card`: текущая карточка с датой, местом, описанием и тегами.
- `Route card`: cover, title, duration, distance, stops, difficulty, author/editorial badge.
- `Route stop`: number, drag handle, thumbnail, visit time, remove action.
- `Map pin`: place / event / route stop / cluster / selected.
- `Route summary`: distance, walking time, visit time, total time.
- `Bottom sheet`: collapsed / half / full.
- `Toast`, `Empty state`, `Skeleton`, `Error state`.
- `Mobile bottom navigation`: Карта / Маршруты / Моя прогулка / Избранное.

## 03 Current product

### D01 — Current map, 1440×1024

Точный перенос текущего интерфейса:

- слева карта 1000 px;
- справа панель 440 px;
- header `Wandero / Историческая карта Петербурга`;
- поиск, период `С/По`, tag select;
- tabs `События / Объекты`;
- status bar и scrollable cards;
- на карте selected card, event points и location markers.

### D02 — Selected event

Показать popup на карте и большую карточку: изображение 16:9, год, место, описание, теги, исторический контекст и media links.

### M01 — Current mobile adaptation, 390×844

Карта во весь экран, search bar сверху, фильтры отдельной кнопкой, список в bottom sheet. Карточка выбранной точки открывается на half-height.

## Концепция прогулок

Главный принцип: пользователь не «редактирует граф маршрута», а собирает понятный список мест. Система сама предлагает оптимальный порядок и сразу показывает время прогулки.

### D10 — Explore + empty walk

- Desktop 1440×1024.
- Справа tabs `Места / Маршруты` вместо исторических технических tabs.
- Над картой floating search и chips: `Архитектура`, `Музеи`, `Парки`, `Необычное`, `Бесплатно`, `С детьми`.
- На каждой карточке `+ В прогулку`.
- Снизу карты компактная sticky-плашка: `Моя прогулка · 0 мест`.

### D11 — Walk builder

- Правая панель переходит в режим `Моя прогулка`.
- Поля: старт `Моё местоположение / выбрать`, финиш `как старт / выбрать`, дата и время.
- Ordered list из 4 остановок с drag handle.
- Действия: удалить, изменить длительность посещения, закрепить порядок.
- На карте зелёная линия и numbered pins.
- Summary: `4,8 км · 1 ч 12 мин пешком · 3 ч 40 мин вместе с остановками`.
- Primary CTA: `Построить прогулку`; secondary: `Сохранить`.
- После перестановки показывать unobtrusive action `Оптимизировать порядок`.

### D12 — Place detail + add

Карточка места поверх карты: фото, статус `Открыто до 18:00`, рейтинг источника, стоимость, рекомендуемое время, доступность, краткая история. Primary CTA `Добавить в прогулку`; после добавления — `Добавлено · 3`.

### D13 — Route ready

Маршрут визуально активен. Панель содержит последовательность остановок, предупреждения о закрытии/ремонте и CTA `Начать прогулку`. Действия: поделиться ссылкой, сохранить, экспортировать GPX.

## Popular routes

### D20 — Route catalog

Hero `Готовые прогулки по Петербургу`, поиск и chips по длительности/району/теме. Grid из route cards:

- `Парадный Петербург` — 2 ч 30 мин, 6 мест;
- `Петроградка: модерн и дворы` — 3 ч, 8 мест;
- `Литературный Петербург` — 2 ч, 7 мест;
- `Острова и набережные` — 3 ч 30 мин, 9 мест;
- `Петербург впервые` — 4 ч, 10 мест;
- `Белые ночи` — 2 ч, вечерний маршрут.

Каждая карточка показывает обложку, polyline-miniature, расстояние, длительность, число остановок и editorial/user badge.

### D21 — Popular route detail

Split layout: карта с линией + editorial story. Список остановок содержит порядковый номер и краткую причину посещения. CTA `Повторить маршрут`; secondary `Изменить под себя` копирует stops в пользовательский builder.

## Mobile-first screens

- `M10 Explore`: карта + search + chips + bottom sheet places.
- `M11 Place detail`: full-height sheet и sticky `Добавить`.
- `M12 My walk`: reorderable stop list и sticky summary.
- `M13 Route ready`: map 55%, next stop sheet 45%.
- `M14 Navigation`: крупная следующая инструкция, расстояние, кнопка `Я на месте`.
- `M15 Popular routes`: vertical feed.
- `M16 Route detail`: cover, metadata, map preview, stops, sticky `Начать`.

## Prototype connections

1. `D10 Place card +` → added state + sticky walk count.
2. Sticky walk → `D11`.
3. `Построить прогулку` → loading overlay → `D13`.
4. `Маршруты` → `D20`; route card → `D21`.
5. `Изменить под себя` → `D11` с предзаполненными stops.
6. Mobile equivalents соединить Smart Animate 250 ms; bottom sheets — Ease Out 300 ms.

## Product decisions to validate

- Начинаем с web/PWA, но проектируем mobile-first: прогулка используется на улице.
- Максимум первой версии — 10–12 stops; это понятно пользователю и совпадает с ограничениями многих optimization API.
- Разделить `время ходьбы`, `время посещений` и `общее время`.
- Не переставлять пользовательские stops молча: оптимизация всегда объяснима и отменяема.
- Исторические события остаются вторым слоем контента внутри place detail, а не главным способом построения прогулки.
- Популярные маршруты сначала редакционные; пользовательские публикации требуют модерации, abuse/reporting и версии маршрута.
