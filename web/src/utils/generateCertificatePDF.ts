import jsPDF from 'jspdf';
import QRCode from 'qrcode';
import type { VerifyData } from '../api/client';
import dayjs from 'dayjs';

const PAGE_W = 210;
const PAGE_H = 297;
const MARGIN = 15;
const CONTENT_W = PAGE_W - 2 * MARGIN;
const LABEL_W = 42;
const BRAND = 'Redmansion \u00B7 OTScan';

function getPublicBaseUrl(): string {
  return import.meta.env.VITE_PUBLIC_URL || window.location.origin;
}

// ─── Helpers ──────────────────────────────────────────────────

function drawBorder(doc: jsPDF) {
  // Outer border
  doc.setDrawColor(33, 50, 91); // #21325b
  doc.setLineWidth(1.2);
  doc.rect(MARGIN - 3, MARGIN - 3, PAGE_W - 2 * (MARGIN - 3), PAGE_H - 2 * (MARGIN - 3));
  // Inner border
  doc.setDrawColor(52, 152, 219); // #3498db
  doc.setLineWidth(0.4);
  doc.rect(MARGIN, MARGIN, CONTENT_W, PAGE_H - 2 * MARGIN);
}

function drawSectionHeader(doc: jsPDF, title: string, y: number): number {
  doc.setFont('helvetica', 'bold');
  doc.setFontSize(11);
  doc.setTextColor(33, 50, 91);
  doc.text(title, MARGIN + 6, y);
  y += 2;
  doc.setDrawColor(52, 152, 219);
  doc.setLineWidth(0.3);
  doc.line(MARGIN + 6, y, PAGE_W - MARGIN - 6, y);
  y += 6;
  return y;
}

function drawRow(doc: jsPDF, label: string, value: string, y: number): number {
  doc.setFont('helvetica', 'bold');
  doc.setFontSize(9);
  doc.setTextColor(100, 100, 100);
  doc.text(label + ':', MARGIN + 8, y);

  doc.setFont('courier', 'normal');
  doc.setFontSize(8);
  doc.setTextColor(40, 40, 40);

  const maxW = CONTENT_W - LABEL_W - 16;
  const lines = doc.splitTextToSize(value, maxW);
  doc.text(lines, MARGIN + 8 + LABEL_W, y);
  return y + Math.max(lines.length, 1) * 4 + 2;
}

function shortHex(hex: string, front = 20, back = 8): string {
  if (hex.length <= front + back + 3) return hex;
  return hex.slice(0, front) + '...' + hex.slice(-back);
}

function parseMerkleProof(proofHex: string) {
  if (!proofHex?.startsWith('0x')) return null;
  const raw = proofHex.slice(2);
  if (raw.length < 136) return null;

  const ruid = '0x' + raw.slice(0, 64);
  const rootHash = '0x' + raw.slice(64, 128);
  const count = parseInt(raw.slice(128, 136), 16);

  const steps: { sibling: string; direction: string }[] = [];
  let offset = 136;
  for (let i = 0; i < count && offset + 66 <= raw.length; i++) {
    const sibling = '0x' + raw.slice(offset, offset + 64);
    const dirByte = parseInt(raw.slice(offset + 64, offset + 66), 16);
    steps.push({ sibling, direction: dirByte === 0 ? 'Left' : 'Right' });
    offset += 66;
  }
  return { ruid, rootHash, steps };
}

// ─── Main ─────────────────────────────────────────────────────

export async function generateCertificatePDF(data: VerifyData): Promise<void> {
  // Generate QR code
  const baseUrl = getPublicBaseUrl();
  const verifyUrl = `${baseUrl}/verify?ruid=${data.ruid}`;
  const qrDataUrl = await QRCode.toDataURL(verifyUrl, { width: 300, margin: 1 });

  const doc = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' });
  let y = MARGIN + 8;

  const needsNewPage = (needed: number) => {
    if (y + needed > PAGE_H - MARGIN - 8) {
      doc.addPage();
      drawBorder(doc);
      y = MARGIN + 8;
    }
  };

  // ── Border ──
  drawBorder(doc);

  // ── Title ──
  doc.setFont('helvetica', 'bold');
  doc.setFontSize(20);
  doc.setTextColor(33, 50, 91);
  doc.text('Certificate of Digital Existence Proof', PAGE_W / 2, y + 4, { align: 'center' });
  y += 12;

  // Decorative line
  doc.setDrawColor(52, 152, 219);
  doc.setLineWidth(0.6);
  doc.line(MARGIN + 20, y, PAGE_W - MARGIN - 20, y);
  y += 8;

  // ── Branding ──
  doc.setFont('helvetica', 'bold');
  doc.setFontSize(14);
  doc.setTextColor(52, 152, 219);
  doc.text(BRAND, PAGE_W / 2, y, { align: 'center' });
  y += 7;

  doc.setFont('helvetica', 'normal');
  doc.setFontSize(9);
  doc.setTextColor(100, 100, 100);
  const preamble =
    'This certificate attests that the following digital identity claim has been cryptographically ' +
    'verified and anchored to the Bitcoin blockchain via OpenTimestamps protocol.';
  const preambleLines = doc.splitTextToSize(preamble, CONTENT_W - 30);
  doc.text(preambleLines, PAGE_W / 2, y, { align: 'center' });
  y += preambleLines.length * 4 + 8;

  // ── Section 1: Bitcoin Anchoring (most important — proves when the claim was made) ──
  if (data.btcBlockHeight && data.btcBlockHeight > 0) {
    y = drawSectionHeader(doc, 'BITCOIN ANCHORING', y);
    y = drawRow(doc, 'BTC Block Height', data.btcBlockHeight.toLocaleString(), y);
    if (data.btcTimestamp && data.btcTimestamp > 0) {
      y = drawRow(doc, 'BTC Timestamp', dayjs.unix(data.btcTimestamp).format('YYYY-MM-DD HH:mm:ss UTC'), y);
    }
    y += 4;
  }

  // ── Section 2: Verification Summary ──
  y = drawSectionHeader(doc, 'VERIFICATION SUMMARY', y);

  y = drawRow(doc, 'RUID', data.ruid, y);
  y = drawRow(doc, 'Status', 'Verified', y);
  if (data.batchID) y = drawRow(doc, 'Batch ID', data.batchID, y);
  if (data.leafCount !== undefined && data.leafCount > 0) {
    y = drawRow(doc, 'Leaf Position', `#${data.leafIndex} of ${data.leafCount.toLocaleString()} RUIDs`, y);
  }
  if (data.claimant) y = drawRow(doc, 'Claimant', data.claimant, y);
  if (data.auid) y = drawRow(doc, 'AUID', data.auid, y);
  if (data.puid) y = drawRow(doc, 'PUID', data.puid, y);
  if (data.submitBlock) y = drawRow(doc, 'Submit Block', data.submitBlock.toLocaleString(), y);
  y += 4;

  // ── Section 2: Cryptographic Proof Chain ──
  needsNewPage(30);
  y = drawSectionHeader(doc, 'CRYPTOGRAPHIC PROOF CHAIN', y);

  if (data.rootHash) y = drawRow(doc, 'Merkle Root', data.rootHash, y);
  if (data.otsDigest) {
    y = drawRow(doc, 'OTS Digest', data.otsDigest, y);
    doc.setFont('helvetica', 'italic');
    doc.setFontSize(7);
    doc.setTextColor(140, 140, 140);
    doc.text('= SHA256(Merkle Root)', MARGIN + 8 + LABEL_W, y - 1);
    y += 3;
  }

  // Merkle Proof Steps
  if (data.merkleProof) {
    const parsed = parseMerkleProof(data.merkleProof);
    if (parsed && parsed.steps.length > 0) {
      needsNewPage(12);
      doc.setFont('helvetica', 'bold');
      doc.setFontSize(9);
      doc.setTextColor(80, 80, 80);
      doc.text('Merkle Proof Steps:', MARGIN + 8, y);
      y += 5;

      for (let i = 0; i < parsed.steps.length; i++) {
        needsNewPage(6);
        const step = parsed.steps[i];
        doc.setFont('courier', 'normal');
        doc.setFontSize(7);
        doc.setTextColor(60, 60, 60);
        doc.text(
          `  Step ${i + 1}: [${step.direction}] ${shortHex(step.sibling, 16, 8)}`,
          MARGIN + 10,
          y,
        );
        y += 4;
      }
      y += 2;
    }
  }

  // OTS Proof Operations
  if (data.parsedOTSProof) {
    const ops = data.parsedOTSProof.operations;
    if (ops.length > 0) {
      needsNewPage(12);
      doc.setFont('helvetica', 'bold');
      doc.setFontSize(9);
      doc.setTextColor(80, 80, 80);
      doc.text('OTS Proof Operations:', MARGIN + 8, y);
      y += 5;

      for (let i = 0; i < ops.length; i++) {
        needsNewPage(6);
        const op = ops[i];
        const opLabel =
          op.op === 'sha256' ? 'SHA-256' :
          op.op === 'ripemd160' ? 'RIPEMD-160' :
          op.op === 'keccak256' ? 'KECCAK-256' :
          op.op.charAt(0).toUpperCase() + op.op.slice(1);
        const detail = op.argument ? `${opLabel} ${shortHex(op.argument, 16, 8)}` : opLabel;

        doc.setFont('courier', 'normal');
        doc.setFontSize(7);
        doc.setTextColor(60, 60, 60);
        doc.text(`  ${i + 1}. ${detail}`, MARGIN + 10, y);
        y += 4;
      }

      // Attestations
      for (const att of data.parsedOTSProof.attestations) {
        needsNewPage(6);
        doc.setFont('courier', 'bold');
        doc.setFontSize(7);
        if (att.type === 'bitcoin') {
          doc.setTextColor(180, 130, 0);
          doc.text(`  => Bitcoin Attestation (Block #${att.btcBlockHeight?.toLocaleString()})`, MARGIN + 10, y);
        } else {
          doc.setTextColor(30, 136, 229);
          doc.text(`  => Pending: ${att.calendarUrl || ''}`, MARGIN + 10, y);
        }
        y += 5;
      }
      y += 2;
    }
  }

  // ── Section 4: Raw OTS Proof Binary ──
  if (data.otsProof) {
    needsNewPage(30);
    y = drawSectionHeader(doc, 'RAW OTS PROOF BINARY', y);

    doc.setFont('helvetica', 'normal');
    doc.setFontSize(7);
    doc.setTextColor(100, 100, 100);
    doc.text('Complete OpenTimestamps proof data (hex-encoded):', MARGIN + 8, y);
    y += 4;

    // Render the raw proof in wrapped monospace lines
    doc.setFont('courier', 'normal');
    doc.setFontSize(6);
    doc.setTextColor(60, 60, 60);

    const proofText = data.otsProof;
    const maxCharsPerLine = 90;
    const maxLines = 20; // Cap to avoid extremely long PDFs
    const totalLines = Math.ceil(proofText.length / maxCharsPerLine);
    const linesToShow = Math.min(totalLines, maxLines);

    for (let i = 0; i < linesToShow; i++) {
      needsNewPage(4);
      const chunk = proofText.slice(i * maxCharsPerLine, (i + 1) * maxCharsPerLine);
      doc.text(chunk, MARGIN + 8, y);
      y += 3.2;
    }

    if (totalLines > maxLines) {
      needsNewPage(6);
      doc.setFont('helvetica', 'italic');
      doc.setFontSize(7);
      doc.setTextColor(140, 140, 140);
      doc.text(
        `... (${proofText.length} characters total, full data available via online verification)`,
        MARGIN + 8,
        y,
      );
      y += 5;
    }
    y += 4;
  }

  // ── QR Code ──
  needsNewPage(55);
  doc.setDrawColor(220, 220, 220);
  doc.setLineWidth(0.2);
  doc.line(MARGIN + 6, y, PAGE_W - MARGIN - 6, y);
  y += 6;

  const qrSize = 38;
  doc.addImage(qrDataUrl, 'PNG', MARGIN + 8, y, qrSize, qrSize);

  doc.setFont('helvetica', 'bold');
  doc.setFontSize(9);
  doc.setTextColor(33, 50, 91);
  doc.text('Scan to verify online:', MARGIN + 8 + qrSize + 8, y + 12);

  doc.setFont('courier', 'normal');
  doc.setFontSize(7);
  doc.setTextColor(52, 152, 219);
  const urlLines = doc.splitTextToSize(verifyUrl, CONTENT_W - qrSize - 24);
  doc.text(urlLines, MARGIN + 8 + qrSize + 8, y + 18);

  doc.setFont('helvetica', 'italic');
  doc.setFontSize(7);
  doc.setTextColor(140, 140, 140);
  doc.text('This QR code links directly to the online verification page.', MARGIN + 8 + qrSize + 8, y + 28);
  y += qrSize + 6;

  // ── Footer (on every page) ──
  const totalPages = doc.getNumberOfPages();
  for (let p = 1; p <= totalPages; p++) {
    doc.setPage(p);
    doc.setDrawColor(52, 152, 219);
    doc.setLineWidth(0.3);
    const footerY = PAGE_H - MARGIN - 10;
    doc.line(MARGIN + 6, footerY, PAGE_W - MARGIN - 6, footerY);

    doc.setFont('helvetica', 'normal');
    doc.setFontSize(7);
    doc.setTextColor(140, 140, 140);
    doc.text(
      `Generated: ${dayjs().format('YYYY-MM-DD HH:mm:ss')} UTC`,
      PAGE_W / 2,
      footerY + 4,
      { align: 'center' },
    );
    doc.text(BRAND + ' Blockchain Explorer', PAGE_W / 2, footerY + 8, { align: 'center' });
  }

  // ── Save ──
  const shortRuid = data.ruid.slice(0, 10);
  doc.save(`OTScan-Certificate-${shortRuid}.pdf`);
}
