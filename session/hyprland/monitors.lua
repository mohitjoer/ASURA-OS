-- Catch-all for outputs without a rule of their own: preferred mode, auto position, scale 1.
hl.monitor({
    output   = "",
    mode     = "preferred",
    position = "auto",
    scale    = 1,
})

-- Explicit fallbacks for virtual machine display connectors (VirtualBox, QEMU, VMware)
-- Prevents Aquamarine from failing with "Monitor has NO FALLBACK MODES, and an INVALID one was requested: 0x0@60.00Hz"
hl.monitor({
    output   = "VGA-1",
    mode     = "1280x800@60",
    position = "auto",
    scale    = 1,
})

hl.monitor({
    output   = "Virtual-1",
    mode     = "1280x800@60",
    position = "auto",
    scale    = 1,
})

hl.monitor({
    output   = "Virtual1",
    mode     = "1280x800@60",
    position = "auto",
    scale    = 1,
})

hl.monitor({
    output   = "default",
    mode     = "1280x800@60",
    position = "auto",
    scale    = 1,
})
