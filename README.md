## gosound

gosound is a go library for loading and playing sounds. It wraps portaudio with
some file parsing logic in a similar vein to the image/\* package(s), so it
should feel familiar.

### TODO

* Compressed (WAVE_FORMAT_EXTENDED) wav files.  See
  http://www-mmsp.ece.mcgill.ca/documents/audioformats/wave/samples.html and
  http://www.music.helsinki.fi/tmt/opetus/uusmedia/esim/index-e.html for
  example files and
  http://www-mmsp.ece.mcgill.ca/Documents/AudioFormats/WAVE/WAVE.html for
  format information
* Other formats:
  * Two FLACC libs to compare:
    * https://github.com/eaburns/flac
    * https://github.com/mewkiz/flac
  * AAC: https://github.com/Comcast/gaad
  * MP2/3/4: https://github.com/tcolgate/mp3
    * For this one, I need to look into the legality of it all before I make
      something public.
  * Vorbis: https://github.com/mccoyst/vorbis
* Streaming interface
* Clean up (and make correct) the Playing() and pause stuffs.
