package config

const (
	RadiusSmall  = 6
	RadiusMedium = 8

	SpacingTiny   = 4
	SpacingSmall  = 6
	SpacingMedium = 8
	SpacingLarge  = 12

	FontSizeSmall  = 11
	FontSizeBody   = 12
	FontSizeMedium = 13
	FontSizeTitle  = 14
	FontSizeLarge  = 16
	FontSizeIcon   = 22

	SpotifyControlSize = 36
)

const StaticThemeCSS = `
	@define-color spark-white rgba(255, 255, 255, 1);
	@define-color spark-white-08 rgba(255, 255, 255, 0.08);
	@define-color spark-white-10 rgba(255, 255, 255, 0.1);
	@define-color spark-white-14 rgba(255, 255, 255, 0.14);
	@define-color spark-white-20 rgba(255, 255, 255, 0.2);
	@define-color spark-white-50 rgba(255, 255, 255, 0.5);
	@define-color spark-white-60 rgba(255, 255, 255, 0.6);
	@define-color spark-white-70 rgba(255, 255, 255, 0.7);
	@define-color spark-white-80 rgba(255, 255, 255, 0.8);
	@define-color spark-black-20 rgba(0, 0, 0, 0.2);
	@define-color spark-black-30 rgba(0, 0, 0, 0.3);
	@define-color spark-black-92 rgba(0, 0, 0, 0.92);
	@define-color spark-spotify-green #1DB954;
`
