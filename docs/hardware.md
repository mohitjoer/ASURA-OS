# Hardware Compatibility & Support Matrix

Toren Linux is built on Void Linux x86_64, runit, Hyprland, and a customized Quickshell desktop.
Hardware support is managed through a layered engine (`generic` &rarr; `vendor` &rarr; `family` &rarr; `model` &rarr; `component`). Fixes are applied narrowly using DMI strings, PCI/USB vendor:device IDs, and loaded kernel modules so that every fix is safe on machines that do not need it.

---

## Support Levels

| Level                | Description                                                                                                                                                                                                                      |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Works**            | Fully supported out of the box using mainline Linux kernel and Void Linux packages. Touchpad gestures, display, audio, power profiles, and function keys work reliably.                                                          |
| **Works with Notes** | Fully operational with mainline kernel, but requires specific user attention (e.g., firmware packages, non-free wireless drivers, or firmware flashing).                                                                         |
| **Partial**          | Core system boots and runs normally (display, keyboard, mouse, networking), but specialized vendor features (e.g. out-of-tree camera stacks or proprietary sensor bridges) require packages not currently shipped by Void Linux. |
| **Not Supported**    | Hardware requiring an out-of-tree patched kernel not available in Void Linux repositories.                                                                                                                                       |

---

## Maker & Family Matrix

### Lenovo

| Family / Line                          | Support Level | Driver & Layer          | Notes                                                                                             |
| -------------------------------------- | ------------- | ----------------------- | ------------------------------------------------------------------------------------------------- |
| **ThinkPad** (T, X, P, E, L, Z series) | **Works**     | `thinkpad_acpi`         | TrackPoint, mute LEDs, battery threshold (start/stop), platform thermal profiles.                 |
| **Legion** (5, 7, Pro, Slim)           | **Works**     | `ideapad_laptop`        | Battery conservation mode (caps charge at 60–80%), Fn+Q thermal profiles (`toren power profile`). |
| **IdeaPad** & **Slim**                 | **Works**     | `ideapad_laptop`        | Battery conservation mode, hotkeys, precision touchpad.                                           |
| **Yoga** (2-in-1 convertibles)         | **Works**     | `ideapad_laptop`, `ymc` | Convertible tablet mode detection, display auto-rotation via `iio-sensor-proxy`.                  |
| **ThinkBook**                          | **Works**     | `ideapad_laptop`        | Battery conservation mode, platform profiles.                                                     |

### Dell

| Family / Line                         | Support Level | Driver & Layer                | Notes                                                                         |
| ------------------------------------- | ------------- | ----------------------------- | ----------------------------------------------------------------------------- |
| **XPS** (13, 14, 15, 16, 17)          | **Works**     | `dell-wmi`, `dell-laptop`     | S2idle/deep sleep tuning, thermal profiles, battery charge end thresholds.    |
| **Latitude** (3000, 5000, 7000, 9000) | **Works**     | `dell-wmi`, `dell-laptop`     | Enterprise power profiles, battery longevity limits, USB-C/Thunderbolt docks. |
| **Inspiron**                          | **Works**     | `dell-wmi`                    | Standard thermal management, keyboard backlight.                              |
| **Precision** Mobile Workstations     | **Works**     | `dell-wmi`, `dell-wmi-sysman` | Workstation sustained cooling modes, dual GPU support.                        |
| **Alienware** (m-series, x-series)    | **Works**     | `alienware-wmi`, `dell-wmi`   | High performance fan boost curves, thermal modes.                             |

### HP (Hewlett-Packard)

| Family / Line                 | Support Level | Driver & Layer | Notes                                                        |
| ----------------------------- | ------------- | -------------- | ------------------------------------------------------------ |
| **EliteBook** & **ProBook**   | **Works**     | `hp-wmi`       | Enterprise thermal profiles, microphone mute LED, hotkeys.   |
| **Spectre** (x360 2-in-1)     | **Works**     | `hp-wmi`       | 2-in-1 orientation sensors, quiet/performance thermal modes. |
| **Envy** (Clamshell & x360)   | **Works**     | `hp-wmi`       | Thermal profiles, privacy camera shutter key.                |
| **Pavilion**                  | **Works**     | `hp-wmi`       | Balanced thermal profiles, keyboard backlight.               |
| **ZBook** Mobile Workstations | **Works**     | `hp-wmi`       | High-load workstation cooling, sustained GPU power delivery. |
| **Omen** & **Victus**         | **Works**     | `hp-wmi`       | Gaming thermal modes, fan control via `toren power profile`. |

### ASUS

| Family / Line                       | Support Level | Driver & Layer            | Notes                                                                   |
| ----------------------------------- | ------------- | ------------------------- | ----------------------------------------------------------------------- |
| **ZenBook**                         | **Works**     | `asus-wmi`                | Quiet and balanced thermal profiles, 80% battery health threshold.      |
| **VivoBook**                        | **Works**     | `asus-wmi`                | Standard thermal management, Fn-lock defaults.                          |
| **ROG** (Strix, Flow, SCAR)         | **Works**     | `asus-wmi`, `asus-nb-wmi` | Gaming thermal modes (quiet, balanced, performance), custom fan curves. |
| **Zephyrus** (G14, G15, G16, M16)   | **Works**     | `asus-wmi`                | Slim chassis thermal tuning, NVIDIA dynamic power management.           |
| **TUF Gaming** (A15, F15, A16, F17) | **Works**     | `asus-wmi`                | High performance cooling profiles, battery charge thresholds.           |
| **ProArt Studiobook**               | **Works**     | `asus-wmi`                | Workstation thermal modes for rendering and content creation.           |

### Acer

| Family / Line                    | Support Level | Driver & Layer | Notes                                                   |
| -------------------------------- | ------------- | -------------- | ------------------------------------------------------- |
| **Nitro** (Nitro 5, Nitro 16/17) | **Works**     | `acer-wmi`     | NitroSense thermal profiles, fan boost.                 |
| **Predator** (Helios, Triton)    | **Works**     | `acer-wmi`     | Predator turbo cooling modes.                           |
| **Swift** (Swift 3, 5, Go, Edge) | **Works**     | `acer-wmi`     | Ultrabook quiet thermal profiles, battery optimization. |
| **Aspire**                       | **Works**     | `acer-wmi`     | Everyday thermal modes, standard keyboard backlight.    |

### Framework

| Family / Line                | Support Level | Driver & Layer                | Notes                                                                       |
| ---------------------------- | ------------- | ----------------------------- | --------------------------------------------------------------------------- |
| **Framework Laptop 13 & 16** | **Works**     | `framework_laptop`, `cros_ec` | Expansion card hotplug, battery charge thresholds, display auto-brightness. |

### MSI (Micro-Star International)

| Family / Line                       | Support Level | Driver & Layer      | Notes                                                   |
| ----------------------------------- | ------------- | ------------------- | ------------------------------------------------------- |
| **Modern** & **Prestige**           | **Works**     | `msi-wmi`, `msi-ec` | Ultrabook thermal modes, battery longevity limits.      |
| **Raider**, **Stealth**, **Katana** | **Works**     | `msi-wmi`, `msi-ec` | Cooler Boost profiles, high-load gaming power delivery. |

### Samsung

| Family / Line                     | Support Level | Driver & Layer       | Notes                                                                   |
| --------------------------------- | ------------- | -------------------- | ----------------------------------------------------------------------- |
| **Galaxy Book** (Pro, 360, Ultra) | **Works**     | `samsung-galaxybook` | Battery Life Extender (caps charge at 85%), silent/optimized fan modes. |

### LG

| Family / Line                | Support Level | Driver & Layer | Notes                                                      |
| ---------------------------- | ------------- | -------------- | ---------------------------------------------------------- |
| **LG Gram** (14, 15, 16, 17) | **Works**     | `lg-laptop`    | Battery Care Limit (80%), silent fan control, reader mode. |

### Huawei & Honor

| Family / Line                | Support Level | Driver & Layer | Notes                                                                  |
| ---------------------------- | ------------- | -------------- | ---------------------------------------------------------------------- |
| **MateBook** & **MagicBook** | **Works**     | `huawei-wmi`   | Battery protection thresholds (40% start / 80% stop), Fn-lock hotkeys. |

### Xiaomi & Redmi

| Family / Line                   | Support Level | Driver & Layer | Notes                                            |
| ------------------------------- | ------------- | -------------- | ------------------------------------------------ |
| **RedmiBook** & **Mi Notebook** | **Works**     | `xiaomi-wmi`   | Fn-lock toggles, keyboard backlight permissions. |

### Razer

| Family / Line                        | Support Level | Driver & Layer | Notes                                                             |
| ------------------------------------ | ------------- | -------------- | ----------------------------------------------------------------- |
| **Razer Blade** (14, 15, 16, 17, 18) | **Works**     | `razerkbd`     | High performance thermal cooling, keyboard backlight permissions. |

### Clevo / ODM (System76, Tuxedo, Schenker, XMG)

| Family / Line     | Support Level | Driver & Layer | Notes                                                      |
| ----------------- | ------------- | -------------- | ---------------------------------------------------------- |
| **Clevo Chassis** | **Works**     | `clevo_wmi`    | Fan curves and thermal profiles for rebranded ODM chassis. |

### Chromebooks

| Family / Line                       | Support Level        | Driver & Layer | Notes                                                                    |
| ----------------------------------- | -------------------- | -------------- | ------------------------------------------------------------------------ |
| **x86 Chromebooks (coreboot/UEFI)** | **Works with Notes** | `coreboot`     | Requires MrChromebox full UEFI ROM flashing. Search key mapped to Super. |

### Microsoft Surface

| Family / Line                   | Support Level | Driver & Layer  | Notes                                                                                                                                                                                                                                                        |
| ------------------------------- | ------------- | --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Surface** (Pro, Laptop, Book) | **Partial**   | Mainline kernel | Mainline boots with working display, keyboard, and Wi-Fi. Advanced features (Surface Aggregator Module on newer revisions, IPTS touchscreen/stylus, camera) require linux-surface kernel patches which Void Linux does not package in official repositories. |

### Apple

| Family / Line                           | Support Level        | Driver & Layer          | Notes                                                                                                                                                                                                                   |
| --------------------------------------- | -------------------- | ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Intel MacBooks (pre-2018, non-T2)**   | **Works with Notes** | `applesmc`, `hid-apple` | Apple SMC fan and keyboard backlight work out of the box. Broadcom Wi-Fi uses open `brcmfmac` or non-free `broadcom-wl`.                                                                                                |
| **Intel MacBooks (2018–2020, T2 chip)** | **Not Supported**    | Mainline kernel         | T2 security chip laptops require an extensive out-of-tree kernel patchset for internal keyboard, trackpad, audio, and Wi-Fi. Void Linux does not ship these patches. External USB input/networking works as a fallback. |
| **Apple Silicon (M1/M2/M3/M4)**         | **Not Supported**    | N/A                     | Toren Linux targets x86_64 architecture.                                                                                                                                                                                |

---

## Component Layers

In addition to maker and family profiles, Toren matches modular components:

### Graphics

- **NVIDIA dGPU**: PCIe RTD3 runtime power management (`D3cold`), `nvidia-drm.modeset=1`, dynamic hybrid GPU switching via `toren gpu switch [hybrid | integrated | discrete]`.
- **AMD dGPU**: Power profile switching, hybrid rendering via `DRI_PRIME=1`.
- **Intel Graphics**: `i915` (GuC/HuC firmware loading) and `xe` driver selection.

### Power & Battery

- **Charge Thresholds**: Per-driver battery longevity limits (ThinkPad, ASUS, Dell, Huawei, LG, Samsung, MSI) via `toren power threshold [percentage]`.
- **Platform Profiles**: ACPI platform profiles (quiet, balanced, performance) via `toren power profile`.
- **Sleep Optimization**: `s2idle` vs `deep` sleep selection via `toren power sleep [s2idle | deep]`.

### Input & Sensors

- **Precision Touchpads**: Optimized palm rejection and clickpad thresholds for I2C HID touchpads.
- **2-in-1 Tablet Mode**: Accelerometer and orientation detection via `iio-sensor-proxy`.
- **Fingerprint Readers**: Goodix, Elan, and Synaptics biometric sensors wired to `toren enable fingerprint` via `fprintd`.

### Audio & Networking

- **Smart Amplifiers**: Firmware calibration paths for Cirrus Logic CS35L41 and Realtek ALC codecs.
- **Sound Open Firmware**: `sof-firmware` for Intel Tiger Lake, Alder Lake, Raptor Lake DSPs.
- **Webcams**: Standard USB webcams supported natively via `uvcvideo`. (Note: Intel IPU6/IPU7 MIPI webcams require out-of-tree userspace not currently in Void).
- **Wi-Fi & Bluetooth**: Firmware loading for Intel (`iwlwifi`), MediaTek (`mt7921`), Realtek (`rtw88`/`rtw89`), Qualcomm (`ath11k`), and Broadcom (`brcm80211`).

---

## Submitting Hardware Reports

Nobody can test every laptop alone, but everyone can test one.

If your laptop is not yet listed or you experience any hardware quirk, submit a report:

```sh
toren hardware report --submit
```

This command:

1. Gathers probe data (DMI, PCI devices, USB devices, loaded modules).
2. Strips all serial numbers, MAC addresses, UUIDs, hostnames, and personal identifiers.
3. Opens a pre-filled GitHub issue in your browser.
4. Leaves the submission completely under your control — nothing is uploaded until you click "Submit new issue".
