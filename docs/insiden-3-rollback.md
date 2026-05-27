# Laporan Insiden 3: Rollback Instan

## Konteks Insiden
Terdapat kasus di mana versi rilis baru mengandung *bug kritis*. Proses membatalkan rilis ini memakan waktu ~25 menit (mulai dari SSH ke server, *stop container*, unduh *image* lama, konfigurasi, dan nyalakan). Ini adalah MTTR (*Mean Time to Recovery*) yang terlalu lambat.

## Solusi Kubernetes
Kubernetes Deployment secara otomatis menyimpan rekam jejak (*history*) dari seluruh versi Pod (ReplicaSet). Mengembalikan aplikasi ke rilis sebelumnya sesederhana mengembalikan konfigurasi (*rollout undo*), dan Kubernetes akan mengeksekusi kebalikan dari proses *rolling update* secara instan.

## Langkah-Langkah Pengujian
Setelah versi aplikasi diperbarui (misal versi 2), kami mendemonstrasikan proses pembatalan/pengembalian paksa ke versi 1 dengan cepat.

1. Jalankan perintah untuk melihat riwayat rilis:
   ```bash
   kubectl rollout history deployment/taskflow-api -n taskflow-prod
   ```
2. Untuk langsung membatalkan ke rilis yang stabil sebelumnya (Rollback), cukup jalankan **satu** baris perintah berikut:
   ```bash
   kubectl rollout undo deployment/taskflow-api -n taskflow-prod
   ```
3. Lakukan verifikasi di aplikasi untuk memastikan bahwa fitur telah kembali ke versi sebelumnya dengan respons lama yang stabil.

## Tabel Perbandingan Efisiensi

| Aspek Penilaian | Penanganan Tradisional/Manual | Menggunakan Kubernetes |
| ----------------- | ------------------------------------------ | ------------------------ |
| **Langkah Kerja** | SSH > matikan container > tarik image lama > start ulang > tes konfig | 1 baris perintah terminal (`kubectl rollout undo`) |
| **Durasi (MTTR)** | ~25 menit | **< 60 detik** |
| **Risiko Human Error** | Tinggi (Terlalu banyak campur tangan *engineer*) | Rendah (Otomatis) |

## Bukti Pengujian
*(Sisipkan Screenshot Terminal saat mengeksekusi `kubectl rollout undo` yang memberikan response `deployment.apps/taskflow-api rolled back`)*

## Hasil
- **Kesimpulan:** Waktu untuk menangani bug kritis dari salah rilis dipangkas drastis menjadi hitungan detik. Insiden 3 diselesaikan.
