# Go TS

Try to parse [TS](https://en.wikipedia.org/wiki/MPEG_transport_stream) file.

## Intro

### Payload

- Program Specific Information ([PSI](https://en.wikipedia.org/wiki/Program-specific_information))
- Packetized Elementary Stream (PES)

#### PSI

```text
PAT (PID 0)
-- program (PID x) PMT 
-- -- stream (PID a)
-- -- stream (PID b)
-- -- ...
-- program (PID y) PMT
... 
```

## Refs

- https://github.com/Comcast/gots
