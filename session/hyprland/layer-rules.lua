hl.layer_rule({
    match = { namespace = "^launcher" },
    no_anim = true,
})
hl.layer_rule({
    match = { namespace = "^launcher$" },
    blur = true,
    ignore_alpha = 0.05,
})
hl.layer_rule({
    match = { namespace = "^toren-notifications$" },
    no_anim = true,
})
hl.layer_rule({
    match = { namespace = "^overview$" },
    blur = true,
    no_anim = true,
})
hl.layer_rule({
    match = { namespace = "^(toren-wallpaper|awww-daemon|toren-livewall|mpvpaper|phonto)$" },
    animation = "fade",
})
hl.layer_rule({
    match = { namespace = "^wallpaper-selector-parallel$" },
    no_anim = true,
    blur = true,
    ignore_alpha = 0.05,
})
hl.layer_rule({
    match = { namespace = "^(quickshell|bar.*|dock.*|toren-.*)$" },
    blur = true,
    ignore_alpha = 0.05,
})
hl.layer_rule({
    match = { namespace = "^quickshell:(regionSelector|screenCorners)$" },
    no_anim = true,
})
hl.layer_rule({
    match = { namespace = "^quickshell:session$" },
    no_anim = true,
    blur = true,
    ignore_alpha = 0.05,
})
