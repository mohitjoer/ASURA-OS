-- Environment for everything the compositor launches. asura-session exports
-- most of this already; it is repeated so a compositor started by hand (a dev
-- checkout) behaves the same. The values must stay equal to the ones there.
hl.env("XCURSOR_THEME", "Adwaita")
hl.env("XCURSOR_SIZE", "24")
hl.env("HYPRCURSOR_THEME", "Adwaita")
hl.env("HYPRCURSOR_SIZE", "24")

hl.env("XDG_CURRENT_DESKTOP", "Asura")
hl.env("XDG_SESSION_DESKTOP", "asura")
hl.env("XDG_SESSION_TYPE", "wayland")

hl.env("GDK_BACKEND", "wayland")
hl.env("MOZ_ENABLE_WAYLAND", "1")
hl.env("QT_QPA_PLATFORM", "wayland")
hl.env("QT_AUTO_SCREEN_SCALE_FACTOR", "1")
hl.env("QT_WAYLAND_DISABLE_WINDOWDECORATION", "1")
hl.env("ELECTRON_OZONE_PLATFORM_HINT", "auto")

-- GTK4 apps (nautilus) hang at startup on wlroots-style compositors when the
-- default renderer asks Mutter's ServiceChannel for a display, which only exists
-- under GNOME. The GL renderer connects directly.
hl.env("GSK_RENDERER", "gl")

hl.env("EDITOR", "nvim")
hl.env("VISUAL", "nvim")
