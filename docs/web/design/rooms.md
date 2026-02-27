# Rooms 页面样式文档

`web/static/rooms.html` 中保留的自定义样式。

| 选择器 | CSS | 用途 |
|--------|-----|------|
| `:root` | `--font-sans: 'Plus Jakarta Sans', ui-sans-serif, system-ui, sans-serif` | 全局字体，body 用 `font-[var(--font-sans)]` |
| `.device-card` | `min-height: 280px; width: 100%; max-width: 280px; min-width: 0; overflow: visible` | 设备卡片/骨架屏尺寸 |
| `.room-devices` | `justify-items: start` | 设备网格对齐 |
| `.card-glow-online` | `box-shadow: 0 0 0 1px rgba(34,197,94,.2), 0 4px 20px -4px rgba(34,197,94,.15)` | 在线卡片边框与阴影 |
| `.card-glow-offline` | `box-shadow: 0 0 0 1px rgba(148,163,184,.15)` | 离线卡片边框 |
| `.toggle-thumb` | `transition: transform 0.25s cubic-bezier(0.4,0,0.2,1)` | 开关滑块位移动画 |
| `.status-pulse` | `animation: status-pulse 2s ease-in-out infinite` | 在线徽章呼吸动画 |
| `@keyframes status-pulse` | `0%,100%{opacity:1} 50%{opacity:.6}` | status-pulse 动画定义 |
| `.skeleton` | `background: linear-gradient(...); animation: shimmer 1.5s infinite` | 骨架屏 shimmer |
| `.dark .skeleton` | `background: linear-gradient(...)` | 暗色模式骨架屏 |
| `@keyframes shimmer` | `to { background-position: 200% 0 }` | shimmer 动画定义 |
| `.btn-press` | `:active { transform: scale(0.98) }` | 按钮点击缩放反馈 |
| `.slider-thumb` | `::-webkit-slider-thumb` / `::-moz-range-thumb` | 亮度/音量滑块 thumb，紫蓝渐变 |
