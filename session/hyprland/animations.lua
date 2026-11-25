hl.config({
    animations = {
        enabled = true,
    },
})

hl.animation({ leaf = "global", enabled = true, speed = 2, bezier = "default" })
hl.animation({ leaf = "workspaces", enabled = true, speed = 2, bezier = "default", style = "slide" })
