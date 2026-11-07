const C = {
  app: '#EDE8DF', panel: '#F9F7F2', card: '#FFFFFF', mutedSurface: '#F3F7F1',
  text: '#17201A', secondary: '#404942', muted: '#6E756F', brand: '#2F6B54',
  historical: '#7A3F20', event: '#C2410C', selected: '#FACC15', border: '#D8D0C4',
  error: '#B42318', blue: '#2C5F9E', river: '#B7D7DD', road: '#D5C9B9'
};

const rgb = hex => {
  const n = parseInt(hex.slice(1), 16);
  return { r: ((n >> 16) & 255) / 255, g: ((n >> 8) & 255) / 255, b: (n & 255) / 255 };
};
const solid = hex => [{ type: 'SOLID', color: rgb(hex) }];
const shadow = [{ type: 'DROP_SHADOW', color: { r: .10, g: .086, b: .067, a: .10 }, offset: { x: 0, y: 12 }, radius: 28, spread: 0, visible: true, blendMode: 'NORMAL' }];
const font = (style = 'Regular') => ({ family: 'Inter', style });

async function loadFonts() {
  for (const style of ['Regular', 'Medium', 'Semi Bold', 'Bold']) {
    try { await figma.loadFontAsync(font(style)); } catch (_) {}
  }
}

function txt(value, size = 14, color = C.text, weight = 'Regular', name = '') {
  const n = figma.createText();
  n.name = name || value.slice(0, 36);
  n.fontName = font(weight);
  n.fontSize = size;
  n.characters = value;
  n.fills = solid(color);
  n.lineHeight = { unit: 'PIXELS', value: Math.round(size * 1.4) };
  n.textAutoResize = 'WIDTH_AND_HEIGHT';
  return n;
}

function frame(name, opts = {}) {
  const n = figma.createFrame();
  n.name = name;
  n.layoutMode = opts.layout || 'VERTICAL';
  if (n.layoutMode !== 'NONE') {
    n.primaryAxisSizingMode = opts.primary || 'AUTO';
    n.counterAxisSizingMode = opts.counter || 'AUTO';
    n.itemSpacing = opts.gap ?? 0;
    const p = opts.padding ?? 0;
    n.paddingTop = opts.pt ?? p; n.paddingRight = opts.pr ?? p;
    n.paddingBottom = opts.pb ?? p; n.paddingLeft = opts.pl ?? p;
  }
  n.fills = opts.fill === null ? [] : solid(opts.fill || C.card);
  if (opts.radius !== undefined) n.cornerRadius = opts.radius;
  if (opts.stroke) { n.strokes = solid(opts.stroke); n.strokeWeight = 1; }
  if (opts.shadow) n.effects = shadow;
  if (opts.clip !== undefined) n.clipsContent = opts.clip;
  if (opts.w) n.resize(opts.w, opts.h || 100);
  if (opts.w && n.layoutMode !== 'NONE') n.counterAxisSizingMode = 'FIXED';
  if (opts.h && n.layoutMode !== 'NONE') n.primaryAxisSizingMode = 'FIXED';
  return n;
}

function rect(name, w, h, fill, radius = 0, stroke = null) {
  const n = figma.createRectangle(); n.name = name; n.resize(w, h); n.fills = solid(fill); n.cornerRadius = radius;
  if (stroke) { n.strokes = solid(stroke); n.strokeWeight = 1; }
  return n;
}

function circle(name, d, fill, stroke = null) {
  const n = figma.createEllipse(); n.name = name; n.resize(d, d); n.fills = solid(fill);
  if (stroke) { n.strokes = solid(stroke); n.strokeWeight = 2; }
  return n;
}

function append(parent, ...children) { children.flat().forEach(c => parent.appendChild(c)); return parent; }
function grow(n) { n.layoutGrow = 1; return n; }
function stretch(n) { n.layoutAlign = 'STRETCH'; return n; }
function label(value, fill = C.mutedSurface, color = C.secondary) {
  const n = frame('Label', { layout: 'HORIZONTAL', gap: 4, padding: 6, pt: 4, pb: 4, fill, radius: 999 });
  append(n, txt(value, 12, color, 'Medium'));
  return n;
}

function button(textValue, type = 'primary', state = 'default', w = null) {
  const fill = state === 'disabled' ? '#C8CDC9' : type === 'primary' ? C.brand : type === 'secondary' ? C.card : C.panel;
  const color = state === 'disabled' ? C.muted : type === 'primary' ? C.card : C.brand;
  const n = frame(`Button/${type}/${state}`, { layout: 'HORIZONTAL', padding: 16, pt: 12, pb: 12, gap: 8, fill, radius: 8, stroke: type === 'secondary' ? C.brand : null });
  if (state === 'loading') append(n, circle('Spinner', 14, fill, color));
  append(n, txt(textValue, 14, color, 'Semi Bold'));
  if (w) { n.resize(w, 44); n.counterAxisSizingMode = 'FIXED'; }
  return n;
}

function chip(value, selected = false) {
  const n = frame(`Filter chip/${selected ? 'Selected' : 'Default'}`, { layout: 'HORIZONTAL', padding: 12, pt: 8, pb: 8, fill: selected ? C.brand : C.card, radius: 999, stroke: selected ? C.brand : C.border });
  append(n, txt(value, 12, selected ? C.card : C.secondary, 'Medium'));
  return n;
}

function search(value = 'Поиск мест и событий', width = 360) {
  const n = frame('Search field', { layout: 'HORIZONTAL', gap: 10, padding: 12, pt: 10, pb: 10, fill: C.card, radius: 8, stroke: C.border, w: width, h: 44 });
  append(n, txt('⌕', 20, C.brand, 'Regular'), grow(txt(value, 14, value.startsWith('Поиск') ? C.muted : C.text)));
  return n;
}

function mapPin(number, x, y, selected = false, parent) {
  const p = frame(`Map pin/${selected ? 'Selected' : 'Route stop'}`, { layout: 'NONE', w: 40, h: 40, fill: selected ? C.selected : C.brand, radius: 20, stroke: C.card, shadow: true });
  const t = txt(String(number), 14, selected ? C.text : C.card, 'Bold'); t.x = number > 9 ? 11 : 15; t.y = 9; p.appendChild(t);
  p.x = x; p.y = y; parent.appendChild(p); return p;
}

function map(name = 'Map', w = 1000, h = 1024, route = true) {
  const m = frame(name, { layout: 'NONE', fill: C.app, w, h, clip: true });
  const river = rect('Neva river', w * .18, h * 1.35, C.river, 90); river.rotation = -22; river.x = w * .46; river.y = -120; m.appendChild(river);
  for (let i = 0; i < 11; i++) {
    const road = rect(`Street ${i + 1}`, w * .9, 3, i % 3 === 0 ? '#C5B7A4' : C.road, 2);
    road.rotation = i % 2 ? -24 : 18; road.x = -80 + (i % 4) * 110; road.y = 70 + i * (h / 11); m.appendChild(road);
  }
  for (let i = 0; i < 7; i++) {
    const park = rect(`Park ${i + 1}`, 100 + i * 8, 62, '#C7D7BF', 18); park.x = 80 + (i * 123) % (w - 160); park.y = 90 + (i * 139) % (h - 180); m.appendChild(park);
  }
  if (route) {
    const line = figma.createLine(); line.name = 'Route line'; line.resize(Math.max(220, w * .55), 0); line.rotation = -18; line.x = w * .20; line.y = h * .56; line.strokes = solid(C.brand); line.strokeWeight = 6; line.strokeCap = 'ROUND'; m.appendChild(line);
    [1, 2, 3, 4].forEach((n, i) => mapPin(n, w * (.22 + i * .16), h * (.64 - i * .08), i === 2, m));
  } else {
    for (let i = 0; i < 8; i++) { const p = circle('Event point', 14, i % 3 ? C.event : C.brand, C.card); p.x = 90 + (i * 109) % (w - 120); p.y = 140 + (i * 103) % (h - 180); m.appendChild(p); }
  }
  return m;
}

function sectionTitle(kicker, title, body = '') {
  const s = frame('Section title', { gap: 6, fill: null });
  append(s, txt(kicker.toUpperCase(), 12, C.brand, 'Semi Bold'), txt(title, 28, C.text, 'Semi Bold'));
  if (body) append(s, txt(body, 14, C.secondary));
  return s;
}

function placeCard(title, subtitle, added = false) {
  const c = frame(`Place card/${added ? 'Added' : 'Default'}`, { layout: 'HORIZONTAL', gap: 12, padding: 12, fill: C.card, radius: 12, stroke: added ? C.brand : C.border, shadow: true });
  append(c, rect('Thumbnail', 80, 80, added ? '#C6D7CE' : '#D9C3A8', 8));
  const info = frame('Content', { gap: 4, fill: null });
  append(info, txt(title, 14, C.text, 'Semi Bold'), txt(subtitle, 12, C.muted), label(added ? 'Добавлено' : 'Открыто до 18:00', added ? C.mutedSurface : '#FFF5D7', added ? C.brand : C.historical));
  append(c, grow(info), button(added ? '✓' : '+', added ? 'secondary' : 'primary'));
  return c;
}

function eventCard(title, year = '1912') {
  const c = frame('Event card', { gap: 8, padding: 16, fill: C.card, radius: 12, stroke: C.border, shadow: true });
  append(c, label(year, '#F5E7DF', C.historical), txt(title, 16, C.text, 'Semi Bold'), txt('Петроградская сторона · исторический контекст и детали события', 12, C.secondary));
  const tags = frame('Tags', { layout: 'HORIZONTAL', gap: 6, fill: null }); append(tags, label('Архитектура'), label('История'));
  append(c, tags); return c;
}

function routeCard(title, time, stops, tint = '#C8D9D0') {
  const c = frame('Route card', { gap: 10, padding: 12, fill: C.card, radius: 12, stroke: C.border, shadow: true, w: 330, h: 300 });
  const cover = frame('Cover + polyline', { layout: 'NONE', fill: tint, radius: 8, w: 306, h: 144, clip: true });
  const l = figma.createLine(); l.resize(190, 0); l.rotation = -18; l.x = 48; l.y = 88; l.strokes = solid(C.brand); l.strokeWeight = 5; l.strokeCap = 'ROUND'; cover.appendChild(l);
  append(cover, label('Выбор редакции', C.card, C.brand)); cover.children[cover.children.length - 1].x = 12; cover.children[cover.children.length - 1].y = 12;
  append(c, cover, txt(title, 18, C.text, 'Semi Bold'), txt(`${time}  ·  ${stops} мест  ·  4,8 км`, 12, C.secondary), txt('Маршрут с понятным темпом и красивыми видами города', 12, C.muted));
  return c;
}

function stopRow(num, title, time = '45 мин') {
  const r = frame('Route stop', { layout: 'HORIZONTAL', gap: 10, padding: 10, fill: C.card, radius: 8, stroke: C.border });
  const badge = frame(`Stop ${num}`, { layout: 'NONE', w: 30, h: 30, fill: C.brand, radius: 15 });
  const numText = txt(String(num), 12, C.card, 'Bold'); numText.x = num > 9 ? 8 : 11; numText.y = 6; badge.appendChild(numText);
  append(r, txt('⠿', 18, C.muted), badge);
  const copy = frame('Stop info', { gap: 2, fill: null }); append(copy, txt(title, 14, C.text, 'Semi Bold'), txt(`Посещение · ${time}`, 12, C.muted));
  append(r, grow(copy), txt('×', 20, C.muted)); return r;
}

function summary() {
  const s = frame('Route summary', { gap: 8, padding: 16, fill: C.mutedSurface, radius: 12, stroke: '#C9D8C8' });
  append(s, txt('Итого', 14, C.text, 'Semi Bold'), txt('4,8 км  ·  1 ч 12 мин пешком', 14, C.brand, 'Semi Bold'), txt('2 ч 28 мин в местах  ·  3 ч 40 мин всего', 12, C.secondary)); return s;
}

function panelHeader(title, subtitle = '') {
  const h = frame('Panel header', { gap: 6, fill: null }); append(h, txt(title, 24, C.text, 'Bold'));
  if (subtitle) append(h, txt(subtitle, 12, C.muted)); return h;
}

function desktopScreen(name, mode = 'explore') {
  const s = frame(name, { layout: 'HORIZONTAL', fill: C.app, w: 1440, h: 1024, clip: true });
  const m = map('Map', 1000, 1024, mode !== 'current'); m.layoutGrow = 1; m.layoutAlign = 'STRETCH';
  const panel = frame('Side panel', { gap: 16, padding: 24, fill: C.panel, w: 440, h: 1024, clip: true });
  panel.counterAxisSizingMode = 'FIXED'; panel.primaryAxisSizingMode = 'FIXED';
  if (mode === 'current') {
    append(panel, panelHeader('Wandero', 'Историческая карта Петербурга'), search(), txt('Период', 12, C.muted, 'Semi Bold'));
    const dates = frame('Date range', { layout: 'HORIZONTAL', gap: 8, fill: null }); append(dates, chip('С 1703'), chip('По 2026')); append(panel, dates);
    const tabs = frame('Tabs', { layout: 'HORIZONTAL', gap: 4, padding: 4, fill: C.app, radius: 8 }); append(tabs, chip('События', true), chip('Объекты')); append(panel, tabs, txt('Найдено 128 событий', 12, C.muted), eventCard('Открытие Каменноостровского театра'), eventCard('Доходный дом в стиле модерн', '1908'), eventCard('Первая трамвайная линия', '1907'));
  } else if (mode === 'builder' || mode === 'ready') {
    append(panel, panelHeader(mode === 'ready' ? 'Прогулка готова' : 'Моя прогулка', 'Петербург · сегодня после 14:00'));
    const fields = frame('Route settings', { gap: 8, padding: 12, fill: C.card, radius: 8, stroke: C.border }); append(fields, txt('Старт  ·  Моё местоположение', 13, C.secondary), txt('Финиш  ·  Как старт', 13, C.secondary), txt('Дата и время  ·  Сегодня, 14:00', 13, C.secondary)); append(panel, fields);
    if (mode === 'ready') append(panel, label('⚠ Михайловский сад закроется в 20:00', '#FFF2E9', C.event));
    ['Исаакиевский собор', 'Новая Голландия', 'Юсуповский дворец', 'Семимостье'].forEach((x, i) => append(panel, stopRow(i + 1, x, i === 1 ? '60 мин' : '35 мин')));
    append(panel, summary(), mode === 'ready' ? button('Начать прогулку', 'primary', 'default', 392) : button('Построить прогулку', 'primary', 'default', 392));
    const acts = frame('Actions', { layout: 'HORIZONTAL', gap: 8, fill: null }); append(acts, button('Сохранить', 'secondary'), button(mode === 'ready' ? 'Поделиться' : 'Оптимизировать порядок', 'secondary')); append(panel, acts);
  } else if (mode === 'detail') {
    append(panel, panelHeader('Особняк Кшесинской', 'Каменноостровский проспект, 2'), rect('Place image', 392, 220, '#CDB9A2', 12), label('Открыто до 18:00', '#EDF7EF', C.brand), txt('4,8 · источник: Яндекс Карты', 12, C.muted), txt('Архитектура северного модерна и одна из самых выразительных городских историй начала XX века.', 14, C.secondary), txt('₽ 400  ·  45–60 мин  ·  Доступная среда', 13, C.text, 'Semi Bold'), button('Добавить в прогулку', 'primary', 'default', 392), eventCard('Исторический слой: события особняка', '1917'));
  } else {
    const searchFloat = frame('Floating search & filters', { layout: 'HORIZONTAL', gap: 8, padding: 12, fill: null }); append(searchFloat, search('Куда пойдём?', 320), chip('Архитектура', true), chip('Музеи'), chip('Парки'), chip('Бесплатно')); searchFloat.x = 24; searchFloat.y = 24; m.appendChild(searchFloat);
    const sticky = frame('My walk sticky', { layout: 'HORIZONTAL', gap: 12, padding: 14, fill: C.text, radius: 12, shadow: true }); append(sticky, txt('Моя прогулка · 0 мест', 14, C.card, 'Semi Bold'), txt('→', 18, C.card)); sticky.x = 350; sticky.y = 944; m.appendChild(sticky);
    append(panel, panelHeader('Исследуйте Петербург', 'Места и готовые маршруты'));
    const tabs = frame('Explore tabs', { layout: 'HORIZONTAL', gap: 4, padding: 4, fill: C.app, radius: 8 }); append(tabs, chip('Места', true), chip('Маршруты')); append(panel, tabs, placeCard('Исаакиевский собор', '1,2 км · 45 мин'), placeCard('Новая Голландия', '2,4 км · 60 мин'), placeCard('Особняк Кшесинской', '3,1 км · 45 мин', true), placeCard('Летний сад', '1,8 км · 40 мин'));
  }
  append(s, m, panel); return s;
}

function mobileScreen(name, mode = 'explore') {
  const s = frame(name, { gap: 0, fill: C.app, w: 390, h: 844, clip: true }); s.primaryAxisSizingMode = 'FIXED'; s.counterAxisSizingMode = 'FIXED';
  if (mode === 'explore' || mode === 'ready') {
    const m = map('Mobile map', 390, mode === 'ready' ? 465 : 440, mode === 'ready'); append(s, m);
    const overlay = frame('Top controls', { gap: 8, padding: 16, fill: null, w: 390, h: 112 }); overlay.layoutPositioning = 'ABSOLUTE'; overlay.x = 0; overlay.y = 0; append(overlay, search('Поиск мест', 358));
    const ch = frame('Chips', { layout: 'HORIZONTAL', gap: 8, fill: null }); append(ch, chip('Места', true), chip('Парки'), chip('Музеи')); append(overlay, ch); s.appendChild(overlay);
  }
  const sheet = frame('Bottom sheet', { gap: 12, padding: 16, fill: C.panel, radius: 20, w: 390, h: mode === 'detail' ? 844 : mode === 'walk' ? 844 : mode === 'navigation' ? 844 : mode === 'ready' ? 379 : 404, clip: true }); sheet.topLeftRadius = 20; sheet.topRightRadius = 20; sheet.bottomLeftRadius = 0; sheet.bottomRightRadius = 0;
  append(sheet, rect('Handle', 40, 4, C.border, 3)); sheet.children[0].layoutAlign = 'CENTER';
  if (mode === 'detail') append(sheet, rect('Photo', 358, 240, '#CDB9A2', 12), panelHeader('Особняк Кшесинской', 'Открыто до 18:00 · 4,8'), txt('Северный модерн, история и городской контекст в одном месте.', 14, C.secondary), summary(), button('Добавить в прогулку', 'primary', 'default', 358));
  else if (mode === 'walk') { append(sheet, panelHeader('Моя прогулка', '4 места · порядок можно менять')); ['Исаакиевский собор', 'Новая Голландия', 'Юсуповский дворец', 'Семимостье'].forEach((x, i) => append(sheet, stopRow(i + 1, x))); append(sheet, summary(), button('Построить', 'primary', 'default', 358)); }
  else if (mode === 'ready') append(sheet, panelHeader('Следующая остановка', 'Исаакиевский собор · 620 м'), txt('Идите прямо по Конногвардейскому бульвару', 16, C.text, 'Semi Bold'), button('Открыть маршрут', 'primary', 'default', 358));
  else if (mode === 'navigation') append(sheet, label('Через 120 м'), txt('Поверните направо', 28, C.text, 'Bold'), txt('на Почтамтскую улицу', 18, C.secondary), rect('Navigation map preview', 358, 360, C.river, 16), button('Я на месте', 'primary', 'default', 358));
  else append(sheet, panelHeader('Рядом с вами', '24 места'), placeCard('Новая Голландия', '12 мин пешком'), placeCard('Исаакиевский собор', '18 мин пешком'));
  append(s, sheet); return s;
}

function currentSelected() {
  const s = desktopScreen('D02 — Selected event', 'current');
  const m = s.children[0];
  const popup = frame('Selected event popup', { gap: 8, padding: 14, fill: C.card, radius: 12, shadow: true, w: 300, h: 210 }); append(popup, label('1912', '#F5E7DF', C.historical), txt('Открытие театра', 18, C.text, 'Semi Bold'), txt('Событие, изменившее культурную карту Петроградской стороны.', 12, C.secondary), button('Подробнее', 'primary')); popup.x = 440; popup.y = 360; m.appendChild(popup);
  return s;
}

function routeCatalog(name = 'D20 — Route catalog', mobile = false) {
  if (mobile) {
    const s = frame(name, { gap: 16, padding: 16, fill: C.panel, w: 390, h: 844, clip: true }); append(s, panelHeader('Готовые прогулки', 'Петербург в удобном темпе'), search('Поиск маршрутов', 358));
    const cs = frame('Filters', { layout: 'HORIZONTAL', gap: 8, fill: null }); append(cs, chip('До 2 часов', true), chip('Центр')); append(s, cs, routeCard('Парадный Петербург', '2 ч 30 мин', 6), routeCard('Белые ночи', '2 ч · вечером', 7, '#C5CEE0')); return s;
  }
  const s = frame(name, { gap: 24, padding: 32, fill: C.panel, w: 1440, h: 1024, clip: true }); append(s, sectionTitle('Популярные маршруты', 'Готовые прогулки по Петербургу', 'Выберите тему и отправляйтесь гулять'), search('Поиск маршрутов', 620));
  const chips = frame('Catalog filters', { layout: 'HORIZONTAL', gap: 8, fill: null }); append(chips, chip('До 2 часов'), chip('2–4 часа', true), chip('Центральный район'), chip('Архитектура')); append(s, chips);
  const grid1 = frame('Route grid row 1', { layout: 'HORIZONTAL', gap: 16, fill: null }); const grid2 = frame('Route grid row 2', { layout: 'HORIZONTAL', gap: 16, fill: null });
  [['Парадный Петербург','2 ч 30 мин',6],['Петроградка: модерн и дворы','3 ч',8],['Литературный Петербург','2 ч',7],['Острова и набережные','3 ч 30 мин',9],['Петербург впервые','4 ч',10],['Белые ночи','2 ч · вечером',7]].forEach((r,i)=>append(i<3?grid1:grid2,routeCard(r[0],r[1],r[2],i%2?'#D8CDBD':'#C8D9D0'))); append(s,grid1,grid2); return s;
}

function routeDetail(name = 'D21 — Popular route detail', mobile = false) {
  if (mobile) {
    const s = frame(name, { gap: 14, padding: 16, fill: C.panel, w: 390, h: 844, clip: true }); append(s, rect('Cover', 358, 220, '#C8D9D0', 16), panelHeader('Петроградка: модерн и дворы', '3 ч · 8 мест · 6,2 км'), txt('Авторская прогулка по самым выразительным адресам района.', 14, C.secondary)); ['Особняк Кшесинской','Австрийская площадь','Дом с башнями'].forEach((x,i)=>append(s,stopRow(i+1,x))); append(s,button('Начать', 'primary','default',358)); return s;
  }
  const s = frame(name, { layout: 'HORIZONTAL', fill: C.app, w: 1440, h: 1024, clip: true }); append(s, map('Editorial route map', 820, 1024, true)); const story=frame('Editorial story',{gap:16,padding:32,fill:C.panel,w:620,h:1024,clip:true}); append(story,sectionTitle('Выбор редакции','Петроградка: модерн и дворы','3 ч · 8 мест · 6,2 км'),txt('Маршрут соединяет известные памятники и тихие дворы, объясняя, как менялся район.',14,C.secondary)); ['Особняк Кшесинской','Австрийская площадь','Дом с башнями','Левашовский хлебозавод'].forEach((x,i)=>append(story,stopRow(i+1,x))); append(story,button('Повторить маршрут','primary','default',556),button('Изменить под себя','secondary','default',556)); append(s,story); return s;
}

function positionGrid(nodes, startX = 80, startY = 120, gapX = 100, gapY = 120) {
  let x = startX, y = startY, rowH = 0;
  for (const n of nodes) { if (x + n.width > 4400) { x = startX; y += rowH + gapY; rowH = 0; } n.x = x; n.y = y; x += n.width + gapX; rowH = Math.max(rowH, n.height); }
}

function makeCover(page) {
  const cover = frame('Wandero — cover', { gap: 24, padding: 64, fill: C.text, w: 1440, h: 900 });
  append(cover, label('PRODUCT DESIGN · WEB / PWA', C.brand, C.card), txt('Городские прогулки\nпо Петербургу', 56, C.card, 'Bold'), txt('Исследуйте места, собирайте понятный список остановок — порядок и время система рассчитает сама.', 22, '#D6DDD8'), txt('Desktop · Mobile · Components · Prototype', 14, '#AEB9B2', 'Semi Bold'));
  const flow = frame('Primary flow', { layout: 'HORIZONTAL', gap: 12, fill: null }); ['Исследовать','Добавить места','Настроить','Построить','Начать прогулку'].forEach((x,i)=>{append(flow,frame(`Step ${i+1}`,{gap:6,padding:16,fill:i===0?C.brand:C.secondary,radius:12}),i<4?txt('→',20,C.card):null); const st=flow.children[flow.children.length-(i<4?2:1)]; append(st,txt(`0${i+1}`,12,C.selected,'Bold'),txt(x,14,C.card,'Semi Bold'));}); append(cover,flow);
  page.appendChild(cover); cover.x=80; cover.y=80;
}

function makeFoundations(page) {
  const f = frame('Foundations', { gap: 32, padding: 48, fill: C.panel, w: 1440, h: 1200 }); append(f,sectionTitle('01 / Foundations','Визуальная система','Inter · шаг 4 px · спокойная городская палитра'));
  append(f,txt('Цветовые токены',18,C.text,'Bold')); const swatches=frame('Color tokens',{layout:'HORIZONTAL',gap:12,fill:null}); Object.entries({Surface_App:C.app,Surface_Panel:C.panel,Surface_Card:C.card,Text_Primary:C.text,Text_Secondary:C.secondary,Brand_Primary:C.brand,Brand_Historical:C.historical,Map_Event:C.event,Map_Selected:C.selected,Border_Default:C.border,State_Error:C.error}).forEach(([k,v])=>{const sw=frame(k,{gap:8,padding:8,fill:C.card,radius:8,stroke:C.border,w:108,h:116}); append(sw,rect('Color',92,60,v,6),txt(k.replace('_','/'),10,C.secondary,'Medium'),txt(v,10,C.muted)); append(swatches,sw);}); append(f,swatches);
  append(f,txt('Типографика',18,C.text,'Bold')); [['Display / 28 / 34',28,'Semi Bold'],['H1 / 24 / 27',24,'Bold'],['H2 / 18 / 23',18,'Bold'],['Body / 14 / 20',14,'Regular'],['Caption / 12 / 16',12,'Regular'],['Button / 14 / 18',14,'Semi Bold']].forEach(x=>append(f,txt(`${x[0]} — Петербург для прогулок`,x[1],C.text,x[2])));
  const geom=frame('Geometry',{layout:'HORIZONTAL',gap:32,fill:null}); const spacing=frame('Spacing',{gap:8,fill:null}); [4,8,12,16,24,32,48,64].forEach(x=>append(spacing,append(frame(`${x}px`,{layout:'HORIZONTAL',gap:8,fill:null}),rect('Scale',x*2,4,C.brand,2),txt(`${x} px`,11,C.muted)))); const radii=frame('Radius',{layout:'HORIZONTAL',gap:12,fill:null}); [8,12,20].forEach(x=>append(radii,append(frame(`${x}px`,{gap:6,fill:null}),rect('Radius sample',80,64,C.app,x,C.brand),txt(`${x} px`,11,C.muted)))); append(geom,spacing,radii); append(f,txt('Отступы и геометрия',18,C.text,'Bold'),geom); page.appendChild(f); f.x=80; f.y=80;
}

function componentFromFrame(n, variantName) {
  const c = figma.createComponent(); c.name = variantName; c.resize(n.width,n.height); c.layoutMode=n.layoutMode; c.primaryAxisSizingMode=n.primaryAxisSizingMode; c.counterAxisSizingMode=n.counterAxisSizingMode; c.itemSpacing=n.itemSpacing; c.paddingTop=n.paddingTop;c.paddingRight=n.paddingRight;c.paddingBottom=n.paddingBottom;c.paddingLeft=n.paddingLeft;c.fills=n.fills;c.cornerRadius=n.cornerRadius;c.strokes=n.strokes;c.strokeWeight=n.strokeWeight;
  while(n.children.length) c.appendChild(n.children[0]); n.remove(); return c;
}

function makeComponents(page) {
  const title=sectionTitle('02 / Components','Компоненты и состояния','Variants и properties для основных продуктовых сценариев'); page.appendChild(title); title.x=80;title.y=60;
  const groups=[];
  const buttonVariants=[]; for(const type of ['Primary','Secondary','Ghost','Icon']) for(const state of ['Default','Hover','Pressed','Disabled','Loading']) { const b=button(type==='Icon'?'＋':'Действие',type.toLowerCase(),state.toLowerCase()); buttonVariants.push(componentFromFrame(b,`Type=${type}, State=${state}`)); }
  const btnSet=figma.combineAsVariants(buttonVariants,page); btnSet.name='Button'; groups.push(btnSet);
  const searchVariants=['Empty','Value','Focused','Error'].map(state=>componentFromFrame(search(state==='Empty'?'Поиск':'Исаакиевский собор',320),`State=${state}`)); const searchSet=figma.combineAsVariants(searchVariants,page);searchSet.name='Search field';groups.push(searchSet);
  const chipVariants=[]; for(const type of ['Category','Duration','Indoor','Free','Family']) for(const state of ['Unselected','Selected']) chipVariants.push(componentFromFrame(chip(type,state==='Selected'),`Type=${type}, State=${state}`)); const chipSet=figma.combineAsVariants(chipVariants,page);chipSet.name='Filter chip';groups.push(chipSet);
  const pinVariants=[]; for(const type of ['Place','Event','Route stop','Cluster','Selected']) { const p=frame('Pin',{layout:'NONE',w:44,h:44,fill:type==='Selected'?C.selected:type==='Event'?C.event:C.brand,radius:22,stroke:C.card}); const t=txt(type==='Cluster'?'12':'1',13,type==='Selected'?C.text:C.card,'Bold');t.x=14;t.y=11;p.appendChild(t);pinVariants.push(componentFromFrame(p,`Type=${type}`)); } const pinSet=figma.combineAsVariants(pinVariants,page);pinSet.name='Map pin';groups.push(pinSet);
  const cards=frame('Cards & navigation',{layout:'HORIZONTAL',gap:20,padding:24,fill:C.panel,radius:16,stroke:C.border}); append(cards,placeCard('Исаакиевский собор','45 мин · 1,2 км'),eventCard('Открытие театра'),routeCard('Парадный Петербург','2 ч 30 мин',6)); groups.push(cards);
  const utilities=frame('Route & feedback',{layout:'HORIZONTAL',gap:20,padding:24,fill:C.panel,radius:16,stroke:C.border}); append(utilities,stopRow(1,'Новая Голландия'),summary(),append(frame('Toast',{layout:'HORIZONTAL',gap:8,padding:14,fill:C.text,radius:8}),txt('✓ Место добавлено',13,C.card,'Semi Bold')));groups.push(utilities);
  positionGrid(groups,80,180,80,100);
}

function createPage(name) { const p=figma.createPage(); p.name=name; return p; }

function addPrototypeLinks(page,d10,d11,d13,d20,d21) {
  const link=(node,dest,duration=.25)=>{try{node.reactions=[{trigger:{type:'ON_CLICK'},actions:[{type:'NODE',destinationId:dest.id,navigation:'NAVIGATE',transition:{type:'SMART_ANIMATE',easing:{type:'EASE_OUT'},duration},preserveScrollPosition:false}]}];}catch(_) {}};
  link(d10,d11); link(d11,d13); link(d20,d21); link(d21,d11);
  page.flowStartingPoints=[{nodeId:d10.id,name:'Desktop explore'}];
}

async function main() {
  await loadFonts();
  const names=['00 Cover & flows','01 Foundations','02 Components','03 Current product','04 Walk planner — desktop','05 Walk planner — mobile','06 Popular routes','07 Prototype'];
  // Never overwrite the user's existing design. Every run creates a separate,
  // timestamped handoff so the complete operation can be undone or removed.
  const runLabel = new Date().toISOString().slice(0, 16).replace('T', ' ');
  const pages=names.map(name=>createPage(`[Wandero ${runLabel}] ${name}`));
  makeCover(pages[0]); makeFoundations(pages[1]); makeComponents(pages[2]);
  const current=[desktopScreen('D01 — Current map','current'),currentSelected(),mobileScreen('M01 — Current mobile adaptation','explore')]; current.forEach(x=>pages[3].appendChild(x)); positionGrid(current);
  const desktop=[desktopScreen('D10 — Explore + empty walk','explore'),desktopScreen('D11 — Walk builder','builder'),desktopScreen('D12 — Place detail + add','detail'),desktopScreen('D13 — Route ready','ready')]; desktop.forEach(x=>pages[4].appendChild(x)); positionGrid(desktop,80,120,100,140);
  const mobile=[mobileScreen('M10 — Explore','explore'),mobileScreen('M11 — Place detail','detail'),mobileScreen('M12 — My walk','walk'),mobileScreen('M13 — Route ready','ready'),mobileScreen('M14 — Navigation','navigation')]; mobile.forEach(x=>pages[5].appendChild(x)); positionGrid(mobile,80,120,60,120);
  const popular=[routeCatalog(),routeDetail(),routeCatalog('M15 — Popular routes',true),routeDetail('M16 — Route detail',true)]; popular.forEach(x=>pages[6].appendChild(x)); positionGrid(popular,80,120,100,140);
  const protoTitle=sectionTitle('07 / Prototype','Основные переходы','Smart Animate 250 ms · bottom sheet Ease Out 300 ms');pages[7].appendChild(protoTitle);protoTitle.x=80;protoTitle.y=60;
  const proto=[desktopScreen('P01 — Explore','explore'),desktopScreen('P02 — Builder','builder'),desktopScreen('P03 — Ready','ready'),routeCatalog('P04 — Routes'),routeDetail('P05 — Route detail')]; proto.forEach(x=>{x.resize(x.width*.55,x.height*.55);pages[7].appendChild(x);}); positionGrid(proto,80,180,80,100); addPrototypeLinks(pages[7],proto[0],proto[1],proto[2],proto[3],proto[4]);
  pages[0].setRelaunchData({open:'Wandero handoff'}); figma.currentPage=pages[0]; figma.viewport.scrollAndZoomIntoView(pages[0].children); figma.notify('Wandero: 8 страниц и 21 экран созданы'); figma.closePlugin('Макет Wandero готов');
}

main().catch(err=>{figma.notify(`Ошибка: ${err.message}`,{error:true,timeout:10000});console.error(err);figma.closePlugin();});
