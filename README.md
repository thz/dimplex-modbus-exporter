# dimplex-modbus-exporter

prometheus exporter for a ethernet modbus module of a dimplex heat pump

## operating status codes

(With the otherwise empty README.md, the following code translation is probably a bit confusing, but I still want to persist it here, as I just found it.)

```
  0: "Aus", "Off"
  1: "Aus", "Off"
  2: "Heizen", "Heating"
  3: "Schwimmbad", "Pool"
  4: "Warmwasser", "Hot water"
  5: "Kuehlen", "Cooling"
  10: "Abtauen", "Defrosting"
  11: "Durchflussueberwachung", "Flow monitoring"
  30: "Sperre", "Lock"
```
