<script lang="ts">
  let installPrompt: Event | null = $state(null);
  let showPrompt = $state(false);
  let dismissed = $state(false);

  $effect(() => {
    // Check if already installed (display-mode: standalone)
    if (window.matchMedia('(display-mode: standalone)').matches) {
      return;
    }

    const handler = () => {
      installPrompt = window.__pwaDeferredPrompt as Event | null;
      if (installPrompt && !dismissed) {
        showPrompt = true;
      }
    };

    const installedHandler = () => {
      showPrompt = false;
      installPrompt = null;
    };

    // Check if prompt is already available
    if (window.__pwaDeferredPrompt) {
      handler();
    }

    window.addEventListener('pwa-install-available', handler);
    window.addEventListener('pwa-installed', installedHandler);

    return () => {
      window.removeEventListener('pwa-install-available', handler);
      window.removeEventListener('pwa-installed', installedHandler);
    };
  });

  async function handleInstall() {
    if (!installPrompt) return;
    (installPrompt as any).prompt();
    const result = await (installPrompt as any).userChoice;
    if (result.outcome === 'accepted') {
      showPrompt = false;
      installPrompt = null;
    }
    window.__pwaDeferredPrompt = null;
  }

  function handleDismiss() {
    showPrompt = false;
    dismissed = true;
    // Re-show after 7 days
    setTimeout(() => { dismissed = false; }, 7 * 24 * 60 * 60 * 1000);
  }
</script>

{#if showPrompt}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="print:hidden fixed left-4 right-4 md:left-auto md:right-4 md:w-96 z-50 bg-white dark:bg-gray-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 p-4 md:p-5 pwa-install-card"
    onclick={(e) => e.stopPropagation()}
    role="dialog"
    aria-label="Instal Aplikasi"
  >
    <div class="flex items-start gap-3">
      <div class="w-10 h-10 bg-[#1A56DB] rounded-xl flex items-center justify-center text-white font-bold text-xs shrink-0 shadow-sm">
        HR
      </div>
      <div class="flex-1 min-w-0">
        <p class="text-sm font-semibold text-gray-900 dark:text-white">Install HRMS</p>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 leading-relaxed">
          Install aplikasi untuk akses cepat dan offline.
        </p>
      </div>
      <button
        onclick={handleDismiss}
        class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer shrink-0"
        aria-label="Tutup"
      >
        <svg class="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    <div class="flex gap-2 mt-4">
      <button
        onclick={handleDismiss}
        class="flex-1 px-3 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-xl transition cursor-pointer"
      >
        Nanti
      </button>
      <button
        onclick={handleInstall}
        class="flex-1 px-3 py-2 text-sm font-semibold text-white bg-[#1A56DB] hover:bg-[#1a47b8] rounded-xl transition cursor-pointer shadow-sm"
      >
        Install
      </button>
    </div>
  </div>
{/if}

<style>
  .pwa-install-card {
    animation: pwa-slide-up 0.3s ease-out;
    bottom: calc(var(--bottom-nav-height, 60px) + env(safe-area-inset-bottom, 0px) + 16px);
  }

  @media (min-width: 768px) {
    .pwa-install-card {
      bottom: 1.5rem;
    }
  }

  @keyframes pwa-slide-up {
    from {
      transform: translateY(100%);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }
</style>
