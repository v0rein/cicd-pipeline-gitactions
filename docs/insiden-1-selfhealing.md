# Laporan Insiden 1: Self-Healing

## Konteks Insiden
Container aplikasi sering crash di tengah malam, dan sistem tidak memiliki mekanisme pemulihan otomatis, menyebabkan downtime berjam-jam hingga ada staf yang menyadarinya di pagi hari.

## Solusi Kubernetes
Menggunakan **Deployment** Kubernetes yang terus mengawasi *state* (kondisi) pod aplikasi. Jika jumlah pod yang sehat berada di bawah batas replika yang ditentukan (dalam hal ini `replicas: 2`), Kubernetes secara instan akan memutar ulang pod pengganti yang baru tanpa intervensi manusia.

## Langkah-Langkah Pengujian
Untuk membuktikan bahwa sistem memiliki kapabilitas *self-healing*, kami melakukan simulasi berikut:

1. Buka dua jendela terminal.
2. Pada **Terminal 1**, pantau status Pod secara real-time dengan perintah:
   ```bash
   kubectl get pods -n taskflow-prod -w
   ```
3. Pada **Terminal 2**, perintahkan Kubernetes untuk menghapus (mensimulasikan crash) salah satu Pod:
   ```bash
   kubectl delete pod <nama-pod-dari-terminal-1> -n taskflow-prod
   ```

## Bukti Pengujian (Log Terminal)
Begitu perintah penghapusan dieksekusi di Terminal 2, Terminal 1 langsung mencatatkan kejadian berikut secara otomatis:
- Pod lama beralih ke status `Terminating`.
- Kubernetes langsung membuat Pod baru dengan status `Pending` > `ContainerCreating`.
- Pod baru beralih ke status `Running`.

*(Sisipkan Screenshot Terminal 1 dan 2 di sini)*

## Hasil
- **Waktu pemulihan (Recovery Time):** < 5 detik.
- **Kesimpulan:** Insiden 1 berhasil diselesaikan. Jika terjadi crash di masa mendatang, layanan akan pulih otomatis dan terhindar dari downtime panjang.
