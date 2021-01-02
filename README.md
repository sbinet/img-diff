# img-diff

`img-diff` is a simple program to display differences between images.

`img-diff` computes the pixel by pixel difference between two images in the NTSC YIQ color space, as described in:

```
  Measuring perceived color difference using YIQ NTSC
  transmission color space in mobile applications.
  Yuriy Kotsarenko, Fernando Ramos.
```

An electronic version is available at:

- http://www.progmat.uaem.mx:8080/artVol2Num2/Articulo3Vol2Num2.pdf

## Example

```
$> img-diff ./testdata/circle-0.png ./testdata/circle-1.png
$> img-diff ./testdata/func-0.png   ./testdata/func-1.png
```
