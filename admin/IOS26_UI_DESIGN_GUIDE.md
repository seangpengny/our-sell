# 🍏 iOS 26 VisionOS Liquid Glass Design System
> **AI Prompt & Implementation Specification Guide**  
> *Copy and paste this entire document into any AI coding assistant to replicate this exact ultra-premium, modern UI design style.*

---

## 🎯 Prompt Instructions for AI
```text
You are an expert Frontend Architect and UI/UX Designer.
Implement the following user interface following the "iOS 26 VisionOS Liquid Glass Design System".
Ensure the design has high visual contrast, silky glassmorphism with backdrop filters, adaptive dual themes (Light & Dark), and micro-animations.
Use the CSS variables, Tailwind classes, and component patterns defined below.
```

---

## 🎨 1. Core Visual Aesthetics & Design Philosophy

* **Vibe:** Ultra-modern Apple iOS 26 / Apple VisionOS interface.
* **Materials:** Double-layer frosted acrylic glass, real specular light reflections, deep OLED dark backgrounds, and subtle blur saturation.
* **Theme Support:** 100% seamless **Light Mode** (Crisp Frosted Glass) and **Dark Mode** (Deep OLED Midnight Glass).
* **Typography:** Modern geometric sans-serif (`Plus Jakarta Sans` or `SF Pro Display`) paired with clean mono numbers (`JetBrains Mono`).

---

## 🛠️ 2. Tailwind CSS v4 & Base CSS Setup

### Google Fonts to Include in `index.html`:
```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600;700&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap" rel="stylesheet">
```

### Complete `main.css` / Global Stylesheet:
```css
@import "tailwindcss";

/* Crucial for Tailwind v4 class-based dark mode toggling */
@custom-variant dark (&:where(.dark, .dark *));

:root {
  --font-sans: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', Roboto, sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
  
  /* iOS 26 Light Mode Tokens */
  --bg-primary: #f1f4f9;
  --bg-surface: rgba(255, 255, 255, 0.85);
  --bg-surface-elevated: rgba(255, 255, 255, 0.96);
  --border-subtle: rgba(0, 0, 0, 0.08);
  --border-specular: rgba(255, 255, 255, 0.9);
  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-muted: #64748b;
  --accent-blue: #0071e3;
  --accent-blue-subtle: rgba(0, 113, 227, 0.12);
  --shadow-glass: 0 16px 36px -10px rgba(0, 0, 0, 0.08), 0 0 0 1px rgba(0, 0, 0, 0.06);
}

html.dark {
  /* iOS 26 Dark OLED Mode Tokens */
  --bg-primary: #080b11;
  --bg-surface: rgba(18, 24, 38, 0.7);
  --bg-surface-elevated: rgba(23, 32, 51, 0.88);
  --border-subtle: rgba(255, 255, 255, 0.09);
  --border-specular: rgba(255, 255, 255, 0.16);
  --text-primary: #f8fafc;
  --text-secondary: #94a3b8;
  --text-muted: #64748b;
  --accent-blue: #2997ff;
  --accent-blue-subtle: rgba(41, 151, 255, 0.16);
  --shadow-glass: 0 25px 50px -12px rgba(0, 0, 0, 0.45), 0 0 0 1px rgba(255, 255, 255, 0.08);
}

body {
  font-family: var(--font-sans);
  background-color: var(--bg-primary);
  color: var(--text-primary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* ==========================================================================
   Liquid Glassmorphism Utility Classes
   ========================================================================== */

/* Standard Card / Container Glass */
.glass-panel {
  background: var(--bg-surface);
  backdrop-filter: blur(28px) saturate(190%);
  -webkit-backdrop-filter: blur(28px) saturate(190%);
  border: 1px solid var(--border-subtle);
  box-shadow: var(--shadow-glass);
}

/* Elevated Popover / Floating Header Glass */
.glass-panel-elevated {
  background: var(--bg-surface-elevated);
  backdrop-filter: blur(32px) saturate(210%);
  -webkit-backdrop-filter: blur(32px) saturate(210%);
  border: 1px solid var(--border-specular);
  box-shadow: var(--shadow-glass);
}

/* Rounded Pill Button / Badge Glass */
.glass-pill {
  background: rgba(0, 0, 0, 0.04);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--border-subtle);
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

html.dark .glass-pill {
  background: rgba(255, 255, 255, 0.08);
}

.glass-pill:hover {
  background: rgba(0, 0, 0, 0.07);
  border-color: var(--border-specular);
  transform: translateY(-1px);
}

html.dark .glass-pill:hover {
  background: rgba(255, 255, 255, 0.14);
}

/* Glass Form Inputs & Textareas */
.glass-input {
  background: rgba(0, 0, 0, 0.03);
  backdrop-filter: blur(16px);
  border: 1px solid var(--border-subtle);
  color: var(--text-primary);
  transition: all 0.2s ease;
}

html.dark .glass-input {
  background: rgba(255, 255, 255, 0.05);
}

.glass-input:focus {
  background: rgba(0, 0, 0, 0.06);
  border-color: var(--accent-blue);
  box-shadow: 0 0 0 3px var(--accent-blue-subtle);
  outline: none;
}

html.dark .glass-input:focus {
  background: rgba(255, 255, 255, 0.09);
}

/* Spring Animations */
@keyframes island-morph {
  0% { transform: scale(0.96); opacity: 0; }
  100% { transform: scale(1); opacity: 1; }
}

.animate-spring-in {
  animation: island-morph 0.35s cubic-bezier(0.175, 0.885, 0.32, 1.15) forwards;
}

@keyframes pulse-subtle {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.75; transform: scale(0.97); }
}

.animate-pulse-subtle {
  animation: pulse-subtle 3s ease-in-out infinite;
}
```

---

## 📱 3. Signature UI Components

### Component A: 🏝️ Dynamic Island Top Header
*Floating, morphing top bar with active workspace/page switcher, live operational pulse, and theme switcher.*

```vue
<!-- DynamicIslandHeader.vue -->
<template>
  <header class="sticky top-0 z-40 px-4 pt-3 pb-2 flex justify-center w-full pointer-events-none">
    <div class="pointer-events-auto glass-panel-elevated rounded-full px-4 py-2 flex items-center justify-between gap-3 md:gap-6 max-w-4xl w-full shadow-2xl transition-all duration-300">
      
      <!-- Left: Logo & Dropdown Switcher -->
      <div class="flex items-center gap-3 relative">
        <div class="w-8 h-8 rounded-full bg-blue-600/20 text-blue-600 dark:text-blue-400 flex items-center justify-center font-bold text-sm shrink-0 ring-1 ring-blue-500/30">
          <Layers :size="16" />
        </div>

        <button
          type="button"
          class="glass-pill px-3 py-1.5 rounded-full flex items-center gap-2.5 text-xs font-semibold hover:border-blue-500/50 transition-all cursor-pointer"
          @click="isDropdownOpen = !isDropdownOpen"
        >
          <div class="w-5 h-5 rounded-full overflow-hidden bg-slate-200 dark:bg-slate-700 ring-1 ring-black/5 dark:ring-white/20 shrink-0">
            <img v-if="activeItem?.avatarUrl" :src="activeItem.avatarUrl" class="w-full h-full object-cover" />
          </div>
          <span class="truncate max-w-28 sm:max-w-48 text-slate-900 dark:text-white font-bold">{{ activeItem?.name }}</span>
          <ChevronDown :size="12" class="opacity-50 transition-transform" :class="{ 'rotate-180': isDropdownOpen }" />
        </button>
      </div>

      <!-- Center: Live Status Indicator (Desktop) -->
      <div class="hidden sm:flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-[11px] font-semibold">
        <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
        <span>Live Operational</span>
      </div>

      <!-- Right: Action Controls (Refresh, Theme Toggle) -->
      <div class="flex items-center gap-1.5">
        <button
          type="button"
          class="w-8 h-8 rounded-full flex items-center justify-center text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-black/5 dark:hover:bg-white/10 transition-colors cursor-pointer"
          @click="toggleTheme"
        >
          <Sun v-if="isDark" :size="14" />
          <Moon v-else :size="14" />
        </button>
      </div>
    </div>
  </header>
</template>
```

---

### Component B: 🧭 VisionOS Segmented Floating Bottom Dock
*Adaptive bottom navigation dock. Uses concise 1-word labels on mobile to avoid truncation and full labels on desktop.*

```vue
<!-- SegmentedNavbar.vue -->
<template>
  <div class="fixed bottom-3 sm:bottom-6 inset-x-0 z-40 flex justify-center px-2 sm:px-4 pointer-events-none">
    <nav class="pointer-events-auto bg-white/85 dark:bg-[#101726]/95 backdrop-blur-2xl p-1.5 rounded-full flex items-center justify-around sm:justify-center gap-1 sm:gap-1.5 shadow-[0_16px_36px_rgba(0,0,0,0.1),0_0_0_1px_rgba(0,0,0,0.08)] dark:shadow-[0_20px_50px_rgba(0,0,0,0.45),0_0_0_1px_rgba(255,255,255,0.12)] border border-black/10 dark:border-white/10 w-full max-w-sm sm:max-w-fit mx-auto transition-all duration-300">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="flex-1 sm:flex-initial flex flex-col sm:flex-row items-center justify-center gap-0.5 sm:gap-2 px-2 sm:px-4 py-1.5 sm:py-2 rounded-full text-[10px] sm:text-xs font-bold transition-all duration-200 cursor-pointer relative select-none group min-w-0"
        :class="activeTab === tab.id
          ? 'bg-blue-600 text-white shadow-lg shadow-blue-500/35 scale-[1.02] ring-1 ring-white/30'
          : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-black/5 dark:hover:bg-white/8'"
        @click="activeTab = tab.id"
      >
        <component
          :is="tab.icon"
          :size="16"
          class="shrink-0 transition-transform group-hover:scale-110"
          :class="activeTab === tab.id ? 'text-white' : 'text-slate-500 dark:text-slate-400 group-hover:text-slate-900 dark:group-hover:text-white'"
        />
        
        <!-- Short Label on Mobile (320px-430px), Full Label on Desktop -->
        <span class="tracking-tight text-[10px] sm:text-xs leading-none mt-0.5 sm:mt-0">
          <span class="sm:hidden">{{ tab.shortLabel }}</span>
          <span class="hidden sm:inline">{{ tab.label }}</span>
        </span>

        <!-- Notification Badge -->
        <span
          v-if="tab.badgeCount && tab.badgeCount > 0"
          class="px-1.5 py-0.5 rounded-full bg-rose-500 text-white text-[9px] font-black"
        >
          {{ tab.badgeCount }}
        </span>
      </button>
    </nav>
  </div>
</template>
```

---

### Component C: 📱 Interactive Side-by-Side Mobile Live Preview Mockup
*Real-time interactive device preview showing what published content looks like before posting.*

```vue
<!-- PostMockup.vue -->
<template>
  <div class="glass-panel-elevated rounded-3xl p-4 sm:p-5 max-w-md w-full mx-auto space-y-3.5 border border-black/5 dark:border-white/10 shadow-2xl">
    <div class="flex items-center justify-between pb-1 border-b border-black/5 dark:border-white/5 text-[11px] text-slate-600 dark:text-slate-400 font-semibold">
      <span class="flex items-center gap-1.5">
        <span class="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
        Live Mobile Preview
      </span>
      <span class="font-mono text-[10px] opacity-60">iOS Mockup</span>
    </div>

    <!-- Authentic Phone Card Wrapper -->
    <div class="bg-white dark:bg-[#18191a] rounded-2xl p-4 border border-black/10 dark:border-white/5 space-y-3 text-slate-900 dark:text-slate-100 shadow-sm transition-colors">
      <!-- Author row -->
      <div class="flex items-center gap-2.5">
        <div class="w-10 h-10 rounded-full overflow-hidden bg-slate-200 dark:bg-slate-700 ring-1 ring-black/5 dark:ring-white/10 shrink-0">
          <img :src="avatarUrl" class="w-full h-full object-cover" />
        </div>
        <div>
          <h4 class="font-bold text-xs text-slate-900 dark:text-white">{{ authorName }}</h4>
          <span class="text-[10px] text-slate-500 dark:text-slate-400">Just now • 🌐 Public</span>
        </div>
      </div>

      <!-- Real-time Body Text -->
      <p class="text-xs leading-relaxed whitespace-pre-line text-slate-800 dark:text-slate-200">
        {{ message || 'Your live content will render here in real-time as you type...' }}
      </p>

      <!-- Action Bar -->
      <div class="pt-2 border-t border-black/5 dark:border-white/5 flex items-center justify-around text-slate-500 dark:text-slate-400 text-xs font-semibold">
        <span class="hover:text-blue-600 cursor-pointer">👍 Like</span>
        <span class="hover:text-blue-600 cursor-pointer">💬 Comment</span>
        <span class="hover:text-blue-600 cursor-pointer">↗️ Share</span>
      </div>
    </div>
  </div>
</template>
```

---

## 🎨 4. Contrast & Theme Safety Rules (DOs & DON'Ts)

| Element | ❌ DON'T (Washed Out / Buggy) | ✅ DO (Adaptive & High Contrast) |
| :--- | :--- | :--- |
| **Headings & Titles** | `class="text-white"` | `class="text-slate-900 dark:text-white font-bold"` |
| **Body & Paragraphs** | `class="text-slate-300"` | `class="text-slate-800 dark:text-slate-200"` |
| **Captions & Meta** | `class="text-slate-400"` | `class="text-slate-500 dark:text-slate-400"` |
| **Card Borders** | `border-white/10` | `border-black/5 dark:border-white/10` |
| **Card Backgrounds** | `bg-slate-900/50` | `bg-white/70 dark:bg-slate-900/50` |
| **Form Inputs** | `bg-transparent text-white` | `class="glass-input text-slate-900 dark:text-white"` |
| **Mobile Nav Tabs** | Multi-word labels without `shortLabel` | Adaptive `shortLabel` (`Pages`, `Studio`, `Inbox`) |

---

## 🚀 5. How to Prompt an AI to use this System

Simply attach or paste this file and add:

> *"Please implement [My Feature / App View] using the iOS 26 VisionOS Liquid Glass design system described above. Use the exact glass utility classes, adaptive light/dark Tailwind tokens, and dynamic floating dock navigation."*
