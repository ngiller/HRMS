<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { notifications, auth } from '$lib/api.js';
  import { onMount } from 'svelte';
  import { slide } from 'svelte/transition';
  import { getAccessibleMenus } from '$lib/permissions.js';
  import { closeMobileMenu, mobileMenuOpen } from '$lib/stores/mobile-nav.svelte.js';

  interface Tab {
    id: string;
    label: string;
    path?: string;
    icon: string;
    isMenu?: boolean;
  }

  const tabs: Tab[] = [
    {
      id: 'dashboard',
      label: 'Beranda',
      path: '/dashboard',
      icon: 'M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6Zm0 9.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6Zm0 9.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z',
    },
    {
      id: 'attendance',
      label: 'Absensi',
      path: '/absensi',
      icon: 'M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5',
    },
    {
      id: 'requests',
      label: 'Pengajuan',
      path: '/cuti',
      icon: 'M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z',
    },
    {
      id: 'payroll',
      label: 'Payroll',
      path: '/penggajian',
      icon: 'M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18',
    },
    {
      id: 'menu',
      label: 'Menu',
      isMenu: true,
      icon: 'M3.75 5.25h16.5m-16.5 4.5h16.5m-16.5 4.5h16.5m-16.5 4.5h16.5',
    },
  ];

  let unreadCount = $state(0);

  let currentPath = $derived($page.url.pathname);

  let activeTab = $derived.by(() => {
    for (const tab of tabs) {
      if (tab.path && currentPath.startsWith(tab.path)) return tab.id;
    }
    return 'dashboard';
  });

  function handleTabClick(tab: Tab) {
    if (tab.isMenu) {
      mobileMenuOpen.value = true;
      return;
    }
    if (tab.path) {
      goto(tab.path);
    }
  }

  async function fetchUnread() {
    try {
      if (!auth.isAuthenticated()) return;
      const res = await notifications.getUnreadCount();
      if (res.success) {
        unreadCount = res.data?.count || 0;
      }
    } catch {
      // silent
    }
  }

  onMount(() => {
    fetchUnread();
    const interval = setInterval(fetchUnread, 60000);
    return () => clearInterval(interval);
  });
</script>

<!-- Bottom Tab Bar — Talenta Style -->
<nav
  class="fixed bottom-0 left-0 right-0 z-50 bg-white/95 dark:bg-gray-900/95 backdrop-blur-lg border-t border-gray-200/70 dark:border-gray-800/70 pb-[env(safe-area-inset-bottom,0px)] shadow-[0_-1px_12px_rgba(0,0,0,0.06)] dark:shadow-[0_-1px_12px_rgba(0,0,0,0.3)]"
  style="height: calc(var(--bottom-nav-height) + env(safe-area-inset-bottom, 0px))"
  aria-label="Navigasi utama"
>
  <div class="flex items-center justify-around h-full px-1 max-w-lg mx-auto">
    {#each tabs as tab}
      {@const isActive = activeTab === tab.id}
      <button
        onclick={() => handleTabClick(tab)}
        class="relative flex flex-col items-center justify-center gap-0.5 w-full h-full rounded-xl transition-all duration-150 cursor-pointer tap-highlight-transparent active:scale-90"
        aria-label={tab.label}
        aria-current={isActive ? 'page' : undefined}
      >
        <div class="relative flex items-center justify-center w-6 h-6">
          {#if tab.isMenu}
            <!-- Grid dots for Menu -->
            <div class="grid grid-cols-3 gap-[2.5px] p-[1px] transition-all duration-200 {isActive ? 'scale-110' : ''}">
              {#each [0, 1, 2, 3, 4, 5, 6, 7, 8] as dot}
                <div
                  class="w-[4.5px] h-[4.5px] rounded-sm transition-colors duration-200"
                  class:bg-[#1A56DB]={isActive}
                  class:bg-gray-400={!isActive}
                  class:dark:bg-blue-400={isActive}
                  class:dark:bg-gray-500={!isActive}
                ></div>
              {/each}
            </div>
          {:else}
            <svg
              class="w-6 h-6 transition-all duration-200 {isActive ? 'scale-110' : ''}"
              fill="{isActive ? '#1A56DB' : 'none'}"
              viewBox="0 0 24 24"
              stroke-width="{isActive ? 1.8 : 1.5}"
              stroke="{isActive ? '#1A56DB' : '#9CA3AF'}"
              class:dark:stroke-blue-400={isActive}
              class:dark:stroke-gray-500={!isActive}
              aria-hidden="true"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d={tab.icon} />
            </svg>
          {/if}
          
          {#if tab.id === 'attendance' && unreadCount > 0}
            <span class="absolute -top-0.5 -right-1.5 w-3.5 h-3.5 bg-red-500 text-white text-[8px] font-bold rounded-full flex items-center justify-center ring-2 ring-white dark:ring-gray-900 shadow-sm">
              {unreadCount > 9 ? '9+' : unreadCount}
            </span>
          {/if}
        </div>
        
        <span
          class="text-[10px] leading-none transition-all duration-200 {isActive ? 'font-semibold text-[#1A56DB] dark:text-blue-400 opacity-100' : 'font-medium text-gray-400 dark:text-gray-500 opacity-70'}"
        >
          {tab.label}
        </span>
        
        <!-- Active dot indicator -->
        {#if isActive && !tab.isMenu}
          <div class="absolute -top-px left-1/2 -translate-x-1/2 w-6 h-[3px] bg-[#1A56DB] dark:bg-blue-400 rounded-full transition-all duration-200"></div>
        {/if}
      </button>
    {/each}
  </div>
</nav>

<!-- Mobile Menu Drawer -->
{#if mobileMenuOpen.value}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    onclick={() => closeMobileMenu()}
    onkeydown={(e) => e.key === 'Escape' && closeMobileMenu()}
    class="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
    role="presentation"
  ></div>
  
  <div
    transition:slide={{ duration: 300 }}
    class="fixed bottom-[var(--bottom-nav-height)] left-0 right-0 z-50 bg-white dark:bg-gray-900 rounded-t-2xl shadow-2xl max-h-[75vh] overflow-y-auto pb-6"
  >
    <!-- Handle bar -->
    <div class="flex justify-center pt-2 pb-1">
      <div class="w-8 h-1 bg-gray-300 dark:bg-gray-600 rounded-full"></div>
    </div>

    <div class="sticky top-0 bg-white dark:bg-gray-900 pb-2 px-5 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
      <span class="text-sm font-bold text-gray-900 dark:text-white">Semua Menu</span>
      <button onclick={() => closeMobileMenu()} class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer" aria-label="Tutup menu">
        <svg class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="px-4 py-3 space-y-4">
      {#each getAccessibleMenus()
        .filter(g => ['Kepegawaian', 'Waktu & Kehadiran', 'Kompensasi & Benefit', 'Informasi'].includes(g.group))
        .sort((a, b) => {
          const order = ['Waktu & Kehadiran', 'Kepegawaian', 'Kompensasi & Benefit', 'Informasi'];
          return order.indexOf(a.group) - order.indexOf(b.group);
        }) as group}
        {#if group.group}
          <div class="text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-widest px-2 pt-2 pb-0.5">{group.group}</div>
        {/if}
        <div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
          {#each group.items as item}
            <a
              href={item.path}
              onclick={() => closeMobileMenu()}
              class="flex flex-col items-center gap-1.5 px-2 py-3 rounded-xl text-center transition {currentPath === item.path ? 'bg-blue-50 dark:bg-[#1A56DB]/10' : 'hover:bg-gray-50 dark:hover:bg-gray-800 active:scale-95 active:bg-gray-100 dark:active:bg-gray-700'}"
            >
              <div class="w-10 h-10 rounded-lg flex items-center justify-center {currentPath === item.path ? 'bg-[#1A56DB] text-white shadow-sm shadow-blue-200' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400'}">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
                </svg>
              </div>
              <span class="text-[10px] leading-tight font-medium text-gray-700 dark:text-gray-300 {currentPath === item.path ? 'text-[#1A56DB] dark:text-blue-400' : ''}">{item.label}</span>
            </a>
          {/each}
        </div>
      {/each}
    </div>
  </div>
{/if}

<style>
  :global(:root) {
    --bottom-nav-height: 60px;
  }
  .tap-highlight-transparent {
    -webkit-tap-highlight-color: transparent;
  }
</style>
