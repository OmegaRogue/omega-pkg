package zerolog_extension

import "github.com/rs/zerolog"

func LoggerWithLevel(logger zerolog.Logger, level zerolog.Level) zerolog.Logger {
	return logger.With().Str(zerolog.LevelFieldName, zerolog.LevelFieldMarshalFunc(level)).Logger()
}
