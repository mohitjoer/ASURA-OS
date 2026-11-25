hl.bind("switch:on:Lid Switch", hl.dsp.exec_cmd("command -v toren-clamshell >/dev/null 2>&1 && toren-clamshell lid close"), { locked = true })
hl.bind("switch:off:Lid Switch", hl.dsp.exec_cmd("command -v toren-clamshell >/dev/null 2>&1 && toren-clamshell lid open"), { locked = true })
hl.exec_cmd("command -v toren-clamshell >/dev/null 2>&1 && toren-clamshell lid sync")
