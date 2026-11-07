# Wandero Figma plugin

Локальный run-once плагин создаёт в текущем Figma Design файле редактируемый Wandero handoff:

- foundations и цветовые токены;
- компоненты с variants;
- desktop- и mobile-экраны;
- сценарии сбора прогулки;
- каталог готовых маршрутов и prototype links.

Плагин не трогает существующие страницы: каждый запуск создаёт новый набор из восьми страниц с меткой времени.

## Первый запуск

1. Откройте Figma Desktop и целевой Figma Design файл.
2. Выберите `Plugins → Development → New plugin… → Figma Design → Run once`.
3. Скопируйте числовой `id` из созданного Figma файла `manifest.json` в `id` нашего `manifest.json`.
4. Выберите `Plugins → Development → Import plugin from manifest…` и укажите этот `manifest.json`.
5. Запустите `Wandero — Walks Handoff Builder` из раздела Development.

Figma официально поддерживает импорт development-плагинов через desktop-приложение. Платный тариф не нужен.

## Повторный запуск

После изменения `code.js` переимпортируйте manifest, затем снова запустите плагин. Если результат не нужен, удалите страницы с одинаковой меткой `[Wandero YYYY-MM-DD HH:MM]`.
