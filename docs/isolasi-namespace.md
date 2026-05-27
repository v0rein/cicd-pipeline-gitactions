# Laporan: Isolasi Namespace (Tugas 6)

## Konteks
Memisahkan lingkungan (*environment*) kerja antara Development (Dev) dan Production (Prod) adalah standar industri untuk menjaga keamanan dan kestabilan sistem.

## Pengujian Isolasi Namespace Kubernetes
Dalam praktiknya, *namespace* memisahkan sistem operasi kluster seakan menjadi "folder" atau "ruang" independen.

1. Kami membuat sebuah Pod sembarang (dummy pod) di namespace `dev` untuk simulasi:
   ```bash
   kubectl run test-pod --image=nginx -n taskflow-dev
   ```
2. Kemudian kami sengaja menciptakan "Kekacauan" (Chaos) dengan memerintahkan penghapusan seluruh pod tanpa ampun di namespace `dev` (layaknya jika developer melakukan uji beban/kegagalan):
   ```bash
   kubectl delete pods --all -n taskflow-dev
   ```
3. Langsung kami verifikasi apakah pod *production* yang sesungguhnya di namespace `taskflow-prod` ikut terseret oleh insiden ini:
   ```bash
   kubectl get pods -n taskflow-prod
   ```

## Bukti Pengujian
*(Sisipkan Screenshot terminal saat menghapus pod di taskflow-dev namun pod di taskflow-prod tetap berstatus Running)*

## Hasil
Kedua namespace **terisolasi secara sempurna**. Tidak ada kebocoran insiden (*blast radius*) lintas namespace.
