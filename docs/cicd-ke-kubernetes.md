# CI/CD ke Kubernetes dengan GitHub Actions

## Diagram Alur
`Push ke GitHub -> GitHub Actions (Build & Test) -> GitHub Packages (GHCR) -> Deploy ke Kubernetes (kubectl set image)`

## Bukti Pengujian
*(Tangkapan layar GitHub Actions CI/CD dan Terminal saat image baru berjalan di Kubernetes)*

## Pertanyaan Evaluasi

**1. Apa yang terjadi di Kubernetes jika job `build` di pipeline gagal? Apakah deployment tetap berjalan?**
Tidak. Jika job `build` gagal, job `deploy` tidak akan dieksekusi (karena ada `needs: build`). Deployment di Kubernetes akan tetap berjalan normal dengan versi sebelumnya tanpa ada gangguan.

**2. Mengapa kita pakai `needs: build` di job `deploy`?**
Untuk memastikan deployment hanya dilakukan jika image baru benar-benar berhasil dibangun, lulus pengujian (tests), dan telah berhasil di-push ke registry. Ini mencegah kita men-deploy versi kode yang rusak.

**3. Apa bedanya pendekatan ini dengan cara deploy manual yang lama?**
Deploy manual mengharuskan developer mengeksekusi banyak langkah secara langsung di server yang rentan error dan lambat. Pendekatan terotomatisasi ini lebih konsisten, aman, cepat, terdokumentasi (history di GitHub), dan tidak memerlukan akses SSH langsung ke server produksi.
