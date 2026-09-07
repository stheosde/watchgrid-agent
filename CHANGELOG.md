# Changelog - Agent

Hier können alle clientseitigen Änderungen von WatchGrid nachgelesen werden.<br>
Unsere Releases orientieren sich an der [semantischen Versionierung](https://semver.org/spec/v2.0.0.html).<br><br>
Das Changelog für Server und Client basiert auf dem System [keepachangelog](https://keepachangelog.com/de/)

## [1.6.0-1.6.1] - 2026-09-07

### Added
- Neuer Überwachungsmechanismus für Docker-Container inkl. automatischer Erkennung laufender Container
- Überprüfung des Container-Status und Übermittlung an das Dashboard
- Eigenständiges Modul `lib/docker` zur Container-Überwachung
- Eigenständiges Modul `lib/services` zur systemd-Überwachung


## [1.5.1-1.5.3] - 2026-08-14

### Fixed
- Beheben von Problematiken in der automatischen Buildpipeline


## [1.5.0] - 2026-08-07

### Changed
- Veröffentlichung des Codes als Open Source unter der GPL-3.0 Lizenz


## [1.4.0] - 2026-07-17

### Added
- Anzahl der CPU, RAM (in GB), Kapazität der Festplatte werden nun an den Server übertragen


## [1.3.0] - 2026-07-10

### Changed
- Änderung des Go Agenten von einer Timersteuerung zu einem Systemd Service


## [1.2.0-1.2.1] - 2026-07-07

### Added
- Lokales Caching der maximalen Uptime

### Fixed
- Starten des Updates nach der Übermittlung von Metriken um falsche Alarme zu vermeiden


## [1.1.0-1.1.2] - 2026-06-29

### Added
- Automatisches Update des Binary bei Releases
- Einführung einer lokalen Konfigurationsdatei. Diese dient als lokaler Cache für den Agent und 
beschreibt die lokal durchzuführenden Tests und zu sammelnden Metriken
- Ports und Services werden nun über die vom Server bereitgetellte Konfiguration überprüft und der
Status übermittelt

### Changed
- Refactoring und Anpassung interner Testfunktionen


## [1.0.0] - 2026-01-23

### Changed
- Refactoring und Trennung von `sysinfo` und `metrics`


## [0.3.0] - 2026-01-21

### Added
- Überprüfung von offenen Ports


## [0.2.0-0.2.2] - 2026-01-10

### Added
- Übermittlung der Agentenversion an den Server

### Changed
- Wiederholter Zustellversuch der Metriken
- Installationsskript angepasst
- Ausgabe der Server Response


## [0.1.0] - 2026-01-07

### Added
- Commandline Flags für Version, Help, Update hinzugefügt

### Changed
- Einlesen der Agent Config über JSON


## [0.0.5] - 2026-01-02

### Added
- Einführung der Clientversionierung
- Automatische Installation über Bash Skript [https://wgri.de/install](https://wgri.de/install)
