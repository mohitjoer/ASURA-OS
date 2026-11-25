hl.window_rule({
    match = { class = ".*" },
    suppress_event = "maximize",
})
hl.window_rule({
    match = {
        class = "^$",
        title = "^$",
        xwayland = true,
        float = true,
        fullscreen = false,
        pin = false,
    },
    no_focus = true,
})
hl.window_rule({
    match = { class = "(pavucontrol|nm-connection-editor|blueman-manager|org.kde.polkit-kde-authentication-agent-1|xdg-desktop-portal-gtk)" },
    float = true,
})
hl.window_rule({
    match = { class = "hyprpolkitagent" },
    float = true,
    center = true,
})
hl.window_rule({
    match = { title = "(Open File|Save File|Save As|Choose Files|Open Folder)" },
    float = true,
})
hl.window_rule({
    match = { class = "[Ss]potify" },
    float = true,
})
hl.window_rule({
    match = { class = "org.gnome.Nautilus" },
    float = true,
    size = "78% 79%",
    center = true,
})
hl.window_rule({
    match = { class = "^([Tt]hunar)$" },
    float = true,
    size = "78% 79%",
    center = true,
})
hl.window_rule({
    match = { title = "^Toren Settings$" },
    float = true,
    size = "100% calc(100% - 33px)",
    move = "0 33",
})
hl.window_rule({
    match = { title = "^(Toren Store)$" },
    float = true,
    size = "61% 70%",
    center = true,
})
hl.window_rule({
    match = { title = "^(asura-vm)$" },
    float = true,
    size = "61% 70%",
    center = true,
    no_anim = true,
})
hl.window_rule({
    match = { class = "asura-store" },
    float = true,
    size = "47% 56%",
    center = true,
})
hl.window_rule({
    match = { class = "looking-glass-client" },
    float = true,
    size = "83% 83%",
    center = true,
})
hl.window_rule({
    match = { class = "[Qq]emu" },
    float = true,
    size = "67% 74%",
    center = true,
})
hl.window_rule({
    match = { title = "^(Welcome to Toren)$" },
    float = true,
    size = "1180 760",
    center = true,
})
hl.window_rule({
    match = { title = "^(Toren Linux Installer)$" },
    float = true,
    size = "960 680",
    center = true,
})
hl.window_rule({
    match = { class = "^(steam|steam_app_.*|gamescope)$" },
    no_blur = true,
    no_shadow = true,
    opaque = true,
    idle_inhibit = "fullscreen",
    immediate = true,
})
hl.window_rule({
    match = { class = "^$", title = "^(.*is sharing.*)$" },
    workspace = "special:sharebar silent",
    no_focus = true,
})
hl.window_rule({
    match = { class = "^([Ss]picy|[Rr]emote-viewer)$" },
    float = true,
    center = true,
})
