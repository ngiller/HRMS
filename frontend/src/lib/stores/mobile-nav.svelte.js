// Svelte 5 runes-based store for mobile navigation state
let _mobileMenuOpen = $state(false);

export const mobileMenuOpen = {
	get value() {
		return _mobileMenuOpen;
	},
	set value(val) {
		_mobileMenuOpen = val;
	}
};

export function closeMobileMenu() {
	_mobileMenuOpen = false;
}

export function openMobileMenu() {
	_mobileMenuOpen = true;
}
