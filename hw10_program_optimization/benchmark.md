# Результаты бенчмарков hw10_program_optimization

## Описание оптимизации

Оптимизирована функция `GetDomainStat` для обработки больших JSON файлов с пользовательскими данными.

### Основные изменения:

1. **Потоковая обработка** - `bufio.Scanner` вместо `io.ReadAll`
2. **Минимальный unmarshaling** - только поле `Email` вместо всей структуры `User`
3. **Работа с байтами** - устранены лишние конвертации string ↔ []byte
4. **Замена regex** - `strings.Contains` вместо `regexp.Match`
5. **Оптимизация строк** - `strings.IndexByte` вместо `strings.SplitN`

## Результаты benchmark тестов

```
goos: darwin
goarch: arm64
pkg: github.com/golang-diasoft-vseroev/hw10_program_optimization

BenchmarkGetDomainStat-11    	      22	 156885119 ns/op	26115787 B/op	  612111 allocs/op
BenchmarkGetDomainStat-11    	      22	 160199928 ns/op	26115901 B/op	  612112 allocs/op
BenchmarkGetDomainStat-11    	      22	 156728276 ns/op	26115858 B/op	  612112 allocs/op
BenchmarkGetDomainStat-11    	      21	 156918298 ns/op	26115951 B/op	  612112 allocs/op
BenchmarkGetDomainStat-11    	      21	 157359788 ns/op	26115858 B/op	  612112 allocs/op

PASS
ok  	github.com/golang-diasoft-vseroev/hw10_program_optimization	21.270s
```

### Среднее значение:
- **Время:** ~157ms на операцию
- **Память:** ~26MB на операцию
- **Аллокации:** ~612,000 аллокаций на операцию

## Результаты теста производительности

```bash
go test -v -count=1 -timeout=30s -tags bench .
```

```
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 155ms / 300ms
    stats_optimization_test.go:47: memory used: 24Mb / 30Mb
--- PASS: TestGetDomainStat_Time_And_Memory (1.91s)
PASS
```

### Соответствие требованиям:
- ✅ **Время выполнения:** 155ms < 300ms (лимит) - **запас 48%**
- ✅ **Использование памяти:** 24MB < 30MB (лимит) - **запас 20%**