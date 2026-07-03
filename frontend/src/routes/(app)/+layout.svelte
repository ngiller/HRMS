<script lang="ts">
  import { page } from "$app/stores";	import 'ag-grid-community/styles/ag-grid.css';
	import 'ag-grid-community/styles/ag-theme-quartz.css';
	import 'leaflet/dist/leaflet.css';
  import { goto } from "$app/navigation";
  import { auth, notifications } from "$lib/api.js";
  import { onMount } from "svelte";
	import { slide } from "svelte/transition";
  import { getAccessibleMenus } from "$lib/permissions.js";

  interface UserData {
    id: string;
    employee_id: string;
    full_name: string;
    email: string;
    role_id: string;
    role_slug: string;
    role_name: string;
    position_name: string;
    department_name: string;
    avatar_initials: string;
  }

  let { children } = $props();
  let sidebarOpen = $state(false);

  // Dark mode — use regular state with init function
  function getInitialDarkMode(): boolean {
    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('hrms_dark_mode');
      if (saved !== null) return saved === 'true';
      return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return false;
  }
  let darkMode = $state(getInitialDarkMode());

  $effect(() => {
    document.documentElement.classList.toggle('dark', darkMode);
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('hrms_dark_mode', String(darkMode));
    }
  });

  // Auth guard: redirect to login if not authenticated
  $effect(() => {
    if (!auth.isAuthenticated()) {
      goto("/login", { replaceState: true });
    }
  });

  let user = $state(auth.getUser() as UserData | null);

  // Re-sync user data (bisa berubah setelah login dari tab lain)
  $effect(() => {
    user = auth.getUser() as UserData | null;
  });

  function getInitials(name: string): string {
    if (!name) return "NA";
    return name.substring(0, 2).toUpperCase();
  }

  async function handleLogout() {
    await auth.logout();
    goto("/login", { replaceState: true });
  }

  // Nav items from permissions helper (filtered by user role)
  let baseNavItems = $derived(getAccessibleMenus());
  
  let navItems = $derived.by(() => {
    if (!menuSearchQuery.trim()) return baseNavItems;
    const query = menuSearchQuery.toLowerCase();
    
    return baseNavItems.map(group => {
      // Filter items matching the query
      const filteredItems = group.items.filter(item => 
        item.label.toLowerCase().includes(query)
      );
      
      // Return new group with filtered items
      return {
        ...group,
        items: filteredItems
      };
    }).filter(group => group.items.length > 0); // Remove empty groups
  });

  function isActive(path: string): boolean {
    return $page.url.pathname === path;
  }

  let dropdownOpen = $state(false);
  let notifDropdownOpen = $state(false);
  let searchQuery = $state("");
  let menuSearchQuery = $state("");
  let unreadCount = $state(0);
  let recentNotifs = $state<any[]>([]);

  async function fetchNotifs() {
    try {
      if (!auth.isAuthenticated()) return;
      const res = await notifications.list(1, 5);
      if (res.success) {
        recentNotifs = res.data.notifications || [];
        unreadCount = res.data.unread_count || 0;
      }
    } catch (e) {
      console.error(e);
    }
  }

  onMount(() => {
    fetchNotifs();
  });

  async function markNotifAsRead(id: string) {
    try {
      await notifications.markAsRead([id]);
      await fetchNotifs();
    } catch (e) {
      console.error(e);
    }
  }
  
  function getNotifIcon(type: string): string {
    if (type.includes("approved") || type === "reimbursement_approved") return "text-emerald-500 bg-emerald-50 dark:bg-emerald-500/10";
    if (type.includes("rejected")) return "text-rose-500 bg-rose-50 dark:bg-rose-500/10";
    if (type === "reprimand_issued") return "text-amber-500 bg-amber-50 dark:bg-amber-500/10";
    return "text-blue-500 bg-blue-50 dark:bg-blue-500/10";
  }

  // Sidebar accordions state
  let openGroups = $state<Record<string, boolean>>({
    "Utama": true,
    "Kepegawaian": false,
    "Waktu & Kehadiran": false,
    "Kompensasi & Benefit": false,
    "Informasi": false,
    "Pengaturan & Master Data": false
  });

  function toggleGroup(group: string) {
    openGroups[group] = !openGroups[group];
  }
  
  $effect(() => {
    const activePath = $page.url.pathname;
    baseNavItems.forEach(group => {
      if (group.group && group.items.some(item => item.path === activePath)) {
        openGroups[group.group] = true;
      }
    });
  });

  function isGroupOpen(groupName: string) {
    if (menuSearchQuery.trim()) return true; // Always open when searching
    return openGroups[groupName];
  }

  function toggleDropdown() {
    dropdownOpen = !dropdownOpen;
  }

  function closeDropdown() {
    dropdownOpen = false;
    notifDropdownOpen = false;
  }

  const menuItems = [
    {
      label: "Profile",
      icon: "M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z",
      href: "/dashboard/profile",
    },
    {
      label: "Change Password",
      icon: "M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z",
      href: "/dashboard/change-password",
    },
    {
      label: "Logout",
      icon: "M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15m-3 0-3-3m0 0 3-3m-3 3H15",
      href: "#",
      onclick: true,
    },
  ];
</script>

<div
  onclick={closeDropdown}
  onkeydown={(e) => e.key === 'Escape' && closeDropdown()}
  role="presentation"
  class="hidden md:flex h-screen overflow-hidden bg-gray-50 dark:bg-gray-950"
>
  <!-- Sidebar -->
  <aside
    class="w-64 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col shrink-0"
  >
    <!-- Logo -->
    <div
      class="h-16 flex items-center gap-3 px-5 border-b border-gray-200 dark:border-gray-800 shrink-0"
    >
      <div
        class="w-9 h-9 bg-[#1A56DB] rounded-lg flex items-center justify-center text-white font-bold text-sm"
      >
        HR
      </div>
      <div>
        <div class="font-bold text-gray-900 dark:text-gray-100 text-sm">HRMS</div>
        <div class="text-xs text-gray-400 dark:text-gray-500">PT Maju Jaya</div>
      </div>
    </div>
      
    <!-- Menu Search Desktop -->
    <div class="px-4 py-3 border-b border-gray-100 dark:border-gray-800">
      <div class="relative">
        <svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          id="menu-search-desktop"
          name="menu_search"
          type="text"
          bind:value={menuSearchQuery}
          placeholder="Cari menu..."
          class="w-full pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition-all text-gray-900 dark:text-gray-100 placeholder:text-gray-400"
        >
      </div>
    </div>

    <!-- Nav Items -->
    <nav class="flex-1 overflow-y-auto py-4 px-3 space-y-1">
      {#each navItems as group}
        <div class="mb-2">
          {#if group.group}
            <button 
              onclick={() => toggleGroup(group.group)}
              class="w-full flex items-center justify-between px-3 py-2 text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider hover:bg-gray-50 dark:hover:bg-gray-800 rounded-lg transition-colors cursor-pointer"
            >
              <span>{group.group}</span>
              <svg 
                class="w-4 h-4 transition-transform duration-200 {isGroupOpen(group.group) ? '' : '-rotate-90'}" 
                fill="none" 
                viewBox="0 0 24 24" 
                stroke-width="2" 
                stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
              </svg>
            </button>
          {/if}
          
          {#if !group.group || isGroupOpen(group.group)}
          <div class="space-y-1 mt-1" transition:slide={{ duration: 200 }}>
            {#each group.items as item}
              <a
                href={item.path}
                class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 group relative {isActive(
                  item.path,
                )
                  ? 'bg-blue-50 dark:bg-[#1A56DB]/10 text-[#1A56DB] dark:text-blue-400 font-semibold'
                  : 'text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white'}"
              >
                <svg
                  class="w-5 h-5 shrink-0"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke-width="1.5"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d={item.icon}
                  />
                </svg>
                <span>{item.label}</span>
                {#if item.badge}
                  <span
                    class="ml-auto bg-red-100 text-red-600 text-xs font-medium px-2 py-0.5 rounded-full"
                    >{item.badge}</span
                  >
                {/if}
              </a>
            {/each}
          </div>
          {/if}
        </div>
      {/each}
    </nav>

    <!-- Bottom section -->
    <div class="p-3 border-t border-gray-200 dark:border-gray-800">
      <a
        href="/dashboard/pengaturan"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-100 transition"
      >
        <svg
          class="w-5 h-5 shrink-0"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z"
          />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"
          />
        </svg>
        <span>Pengaturan</span>
      </a>
    </div>
  </aside>

  <!-- Right Side: Topbar + Content -->
  <div class="flex-1 flex flex-col min-w-0">
    <!-- Topbar -->
    <header
      class="h-16 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between px-6 shrink-0"
    >
      <!-- Left: Search -->
      <div class="relative">
        <svg
          class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="2"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"
          />
        </svg>
        <input
          id="search-karyawan"
          type="search"
          bind:value={searchQuery}
          class="w-72 pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:focus:bg-gray-800 transition placeholder:text-gray-400 dark:placeholder:text-gray-500 text-gray-900 dark:text-gray-100"
          placeholder="Cari karyawan..."
        />
      </div>

      <!-- Right: Notification + User -->
      <div class="flex items-center gap-4">
        <!-- Dark Mode Toggle -->
        <button
          onclick={() => (darkMode = !darkMode)}
          class="relative p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer"
          aria-label={darkMode ? 'Mode Terang' : 'Mode Gelap'}
        >
          {#if darkMode}
            <svg class="w-5 h-5 text-amber-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z" />
            </svg>
          {:else}
            <svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" />
            </svg>
          {/if}
        </button>

        <!-- Notification Bell -->
        <div class="relative">
          <button
            onclick={(e) => { e.stopPropagation(); notifDropdownOpen = !notifDropdownOpen; dropdownOpen = false; }}
            class="relative p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer"
            aria-label="Notifikasi"
          >
            <svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0" />
            </svg>
            {#if unreadCount > 0}
              <span class="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full border border-white dark:border-gray-900"></span>
            {/if}
          </button>

          {#if notifDropdownOpen}
            <div 
              transition:slide={{ duration: 200 }}
              onclick={(e) => e.stopPropagation()}
              onkeydown={(e) => { if (e.key === "Escape") notifDropdownOpen = false; }}
              role="menu"
              tabindex="-1"
              class="absolute right-0 mt-2 w-80 bg-white dark:bg-gray-800 rounded-xl shadow-xl border border-gray-100 dark:border-gray-700 overflow-hidden z-50 flex flex-col max-h-[28rem]"
            >
              <div class="p-3 border-b border-gray-100 dark:border-gray-700 flex justify-between items-center bg-gray-50/50 dark:bg-gray-800/50">
                <span class="font-semibold text-gray-900 dark:text-white text-sm">Notifikasi Baru</span>
                {#if unreadCount > 0}
                  <span class="text-xs font-medium bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400 px-2 py-0.5 rounded-full">{unreadCount}</span>
                {/if}
              </div>
              
              <div class="overflow-y-auto flex-1 divide-y divide-gray-50 dark:divide-gray-700/50">
                {#if recentNotifs.length === 0}
                  <div class="p-6 text-center text-gray-500 dark:text-gray-400 text-sm">Belum ada notifikasi baru.</div>
                {:else}
                  {#each recentNotifs as notif}
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div 
                      onclick={() => { 
                        if (!notif.is_read) markNotifAsRead(notif.id);
                        notifDropdownOpen = false;
                        goto('/notifikasi');
                      }}
                      class="p-3 flex gap-3 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition cursor-pointer relative {notif.is_read ? 'opacity-75' : 'bg-blue-50/20 dark:bg-blue-900/10'}"
                    >
                      {#if !notif.is_read}
                        <div class="absolute left-0 top-0 bottom-0 w-0.5 bg-blue-500"></div>
                      {/if}
                      <div class="w-8 h-8 rounded-full shrink-0 flex items-center justify-center mt-0.5 {getNotifIcon(notif.notification_type)}">
                        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                           <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                        </svg>
                      </div>
                      <div class="min-w-0 flex-1">
                        <p class="text-sm font-medium text-gray-900 dark:text-white truncate">{notif.title}</p>
                        <p class="text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5">{notif.body}</p>
                      </div>
                    </div>
                  {/each}
                {/if}
              </div>
              
              <div class="p-2 border-t border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
                <a 
                  href="/notifikasi" 
                  onclick={() => { notifDropdownOpen = false; }}
                  class="block w-full text-center py-2 text-xs font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 hover:bg-blue-50 dark:hover:bg-gray-700 rounded-lg transition"
                >
                  Lihat Semua Notifikasi
                </a>
              </div>
            </div>
          {/if}
        </div>

        <!-- User Avatar + Dropdown -->
        <div class="relative">
          <button
            onclick={(e) => {
              e.stopPropagation();
              toggleDropdown();
            }}
            class="flex items-center gap-3 pl-4 border-l border-gray-200 cursor-pointer group"
            aria-label="Menu pengguna"
            aria-expanded={dropdownOpen}
          >
            <div
              class="w-9 h-9 bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] rounded-full flex items-center justify-center text-white text-xs font-semibold shrink-0"
            >
              {user?.avatar_initials ||
                getInitials(user?.full_name || "") ||
                "NA"}
            </div>
            <div class="hidden sm:block text-left">
              <div
                class="text-sm font-medium text-gray-900 dark:text-gray-100 group-hover:text-[#1A56DB] dark:group-hover:text-blue-400 transition"
              >
                {user?.full_name || "Pengguna"}
              </div>
              <div class="text-xs text-gray-400">
                {user?.position_name || user?.role_name || ""}
              </div>
            </div>
            <svg
              class="w-4 h-4 text-gray-400 hidden sm:block transition {dropdownOpen
                ? 'rotate-180'
                : ''}"
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="2"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="m19.5 8.25-7.5 7.5-7.5-7.5"
              />
            </svg>
          </button>

          <!-- Dropdown Menu -->
          {#if dropdownOpen}
             <div
              onclick={(e) => e.stopPropagation()}
              onkeydown={(e) => { if (e.key === 'Enter') e.stopPropagation(); }}
              role="menu"
              tabindex="-1"
              class="absolute right-0 top-full mt-2 w-56 bg-white dark:bg-gray-800 rounded-xl shadow-lg dark:shadow-gray-900/50 border border-gray-200 dark:border-gray-700 py-1.5 z-50"
            >
              {#each menuItems as item}
                {#if item.label === "Logout"}
                  <button
                    onclick={handleLogout}
                    class="flex items-center gap-3 w-full text-left px-4 py-2.5 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 transition"
                    role="menuitem"
                  >
                    <svg
                      class="w-4 h-4 shrink-0"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke-width="1.5"
                      stroke="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d={item.icon}
                      />
                    </svg>
                    {item.label}
                  </button>
                {:else}
                  <a
                    href={item.href}
                    class="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition"
                    role="menuitem"
                  >
                    <svg
                      class="w-4 h-4 shrink-0"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke-width="1.5"
                      stroke="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d={item.icon}
                      />
                    </svg>
                    {item.label}
                  </a>
                {/if}
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 overflow-y-auto">
      <div class="p-6">
        {@render children()}
      </div>
    </main>
  </div>
</div>

<!-- Mobile Layout -->  <div class="md:hidden min-h-screen bg-gray-50 dark:bg-gray-950 flex flex-col">
  <!-- Mobile Top Bar -->
  <div
    class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-4 py-3 flex items-center justify-between shrink-0"
  >
    <button
      onclick={() => (sidebarOpen = !sidebarOpen)}
      class="p-2 -ml-2 rounded-lg hover:bg-gray-100 transition cursor-pointer"
      aria-label="Buka menu"
    >
      <svg
        class="w-5 h-5 text-gray-600"
        fill="none"
        viewBox="0 0 24 24"
        stroke-width="1.5"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
        />
      </svg>
    </button>
    <div class="flex items-center gap-2">
      <div
        class="w-7 h-7 bg-[#1A56DB] rounded-lg flex items-center justify-center text-white font-bold text-xs"
      >
        HR
      </div>
      <span class="font-bold text-sm text-gray-900">HRMS</span>
    </div>
    <!-- Mobile User Avatar -->
    <div class="relative">
      <button
        onclick={(e) => {
          e.stopPropagation();
          toggleDropdown();
        }}
        class="w-8 h-8 bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] rounded-full flex items-center justify-center text-xs font-semibold text-white cursor-pointer"
        aria-label="Menu pengguna"
      >
        {user?.avatar_initials || getInitials(user?.full_name || "") || "NA"}
      </button>
          {#if dropdownOpen}
        <div
          transition:slide={{ duration: 200 }}
          onclick={(e) => e.stopPropagation()}
          onkeydown={(e) => { if (e.key === 'Enter') e.stopPropagation(); }}
          role="menu"
          tabindex="-1"
          class="absolute right-0 top-full mt-2 w-48 bg-white rounded-xl shadow-lg border border-gray-200 py-1.5 z-50"
        >
          {#each menuItems as item}
            {#if item.label === "Logout"}
              <button
                onclick={handleLogout}
                class="flex items-center gap-3 w-full text-left px-4 py-2.5 text-sm text-red-600 hover:bg-red-50 transition"
                role="menuitem"
              >
                <svg
                  class="w-4 h-4 shrink-0"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke-width="1.5"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d={item.icon}
                  />
                </svg>
                {item.label}
              </button>
            {:else}
              <a
                href={item.href}
                class="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition"
                role="menuitem"
              >
                <svg
                  class="w-4 h-4 shrink-0"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke-width="1.5"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d={item.icon}
                  />
                </svg>
                {item.label}
              </a>
            {/if}
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <!-- Mobile Sidebar Overlay -->
  {#if sidebarOpen}
    <div class="fixed inset-0 z-50 flex">
      <div
        onclick={() => (sidebarOpen = false)}
        onkeydown={(e) => e.key === "Escape" && (sidebarOpen = false)}
        class="absolute inset-0 bg-black/40 transition-opacity"
        role="presentation"
      ></div>
      <aside class="relative w-72 bg-white h-full shadow-xl flex flex-col">
        <div class="h-16 flex items-center gap-3 px-5 border-b border-gray-200">
          <div
            class="w-9 h-9 bg-[#1A56DB] rounded-lg flex items-center justify-center text-white font-bold text-sm"
          >
            HR
          </div>
          <div>
            <div class="font-bold text-gray-900 text-sm">HRMS</div>
            <div class="text-xs text-gray-400">PT Maju Jaya</div>
          </div>
        </div>
        
        <!-- Menu Search Mobile -->
        <div class="px-4 py-3 border-b border-gray-100">
          <div class="relative">
            <svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              id="menu-search-mobile"
              name="menu_search"
              type="text"
              bind:value={menuSearchQuery}
              placeholder="Cari menu..."
              class="w-full pl-9 pr-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition-all text-gray-900 placeholder:text-gray-400"
            >
          </div>
        </div>

        <nav class="flex-1 overflow-y-auto p-3 space-y-1">
          {#each navItems as group}
            <div class="mb-1">
              {#if group.group}
                <button 
                  onclick={() => toggleGroup(group.group)}
                  class="w-full flex items-center justify-between px-3 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider hover:bg-gray-50 rounded-lg transition-colors"
                >
                  <span>{group.group}</span>
                  <svg 
                    class="w-4 h-4 transition-transform {isGroupOpen(group.group) ? '' : '-rotate-90'}" 
                    fill="none" 
                    viewBox="0 0 24 24" 
                    stroke-width="2" 
                    stroke="currentColor"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                  </svg>
                </button>
              {/if}
              
              {#if !group.group || isGroupOpen(group.group)}
              <div class="space-y-0.5 mt-1" transition:slide={{ duration: 200 }}>
                {#each group.items as item}
                  <a
                    href={item.path}
                    onclick={() => (sidebarOpen = false)}
                    class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition {isActive(
                      item.path,
                    )
                      ? 'bg-blue-50 text-[#1A56DB] font-semibold'
                      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
                  >
                    <svg
                      class="w-5 h-5 shrink-0"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke-width="1.5"
                      stroke="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d={item.icon}
                      />
                    </svg>
                    {item.label}
                  </a>
                {/each}
              </div>
              {/if}
            </div>
          {/each}
        </nav>
      </aside>
    </div>
  {/if}

  <!-- Mobile Content -->
  <main class="flex-1 overflow-y-auto p-4">
    {@render children()}
  </main>
</div>
