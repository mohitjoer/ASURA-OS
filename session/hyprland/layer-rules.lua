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
    match = { namespace = "^asura-notifications$" },
    no_anim = true,
})
hl.layer_rule({
    match = { namespace = "^overview$" },
    blur = true,
    no_anim = true,
})
hl.layer_rule({
    match = { namespace = "^(asura-wallpaper|awww-daemon|asura-livewall|mpvpaper|phonto)$" },
    animation = "fade",
})
hl.layer_rule({
    match = { namespace = "^wallpaper-selector-parallel$" },
    no_anim = true,
    blur = true,
    ignore_alpha = 0.05,
})
hl.layer_rule({
    match = { namespace = "^(quickshell|bar.*|dock.*|asura-.*)$" },
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
