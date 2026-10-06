<p align="center">
  <img src="cmd/iconconvert/winres/icon.svg" width="96" alt="IconConvert">
</p>

<h1 align="center">IconConvert</h1>

<p align="center">Converte imagens em <code>.ico</code> de 16 a 256 px, direto no Windows.</p>

## Uso

- **Arrastar:** solte uma ou mais imagens sobre o `IconConvert.exe`.
- **Terminal:** com o exe no `PATH`, rode `IconConvert logo.png`.

O `.ico` é criado ao lado da imagem: `favicon.ico` para uma imagem, ou o nome de cada imagem quando são várias. Nada é sobrescrito; se o nome existir, vira `favicon2.ico`.

**Formatos:** PNG, JPEG, GIF, BMP, TIFF, WebP, ICO, AVIF, HEIC e SVG.

## Build

```powershell
make build   # bin\IconConvert.exe
```

<p align="center">© 2026 Majestic World Studio</p>
