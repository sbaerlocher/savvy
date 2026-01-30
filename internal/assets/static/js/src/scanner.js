/**
 * Barcode Scanner Module (html5-qrcode)
 * Provides scanning functionality for Cards, Vouchers, and Gift Cards
 * with automatic barcode type detection
 */

// Barcode format mapping from html5-qrcode to internal format names
const BARCODE_TYPE_MAPPING = {
  QR_CODE: 'QR',
  AZTEC: 'AZTEC',
  DATA_MATRIX: 'DATAMATRIX',
  PDF_417: 'PDF417',
  MAXICODE: 'MAXICODE',
  CODE_128: 'CODE128',
  CODE_39: 'CODE39',
  CODE_93: 'CODE93',
  CODABAR: 'CODABAR',
  EAN_13: 'EAN13',
  EAN_8: 'EAN8',
  UPC_A: 'UPCA',
  UPC_E: 'UPCE',
  ITF: 'ITF',
  ITF_14: 'ITF14'
}

/**
 * Create a barcode scanner with automatic type detection
 * @param {Object} config - Scanner configuration
 * @param {Array<number>} formats - Supported barcode formats
 * @returns {Object} Alpine.js component data
 */
function createBarcodeScanner (config, formats = null) {
  return {
    isNewMerchant: false,
    [config.fieldName]: config.initialValue,
    barcodeType: '',
    scanning: false,
    scanMessage: config.defaultMessage,
    html5QrCode: null,

    /**
     * Get supported barcode formats
     */
    getSupportedFormats () {
      if (formats && formats.length > 0) {
        return formats
      }

      // All supported formats if none specified
      return [
        // 2D Barcodes
        Html5QrcodeSupportedFormats.QR_CODE,
        Html5QrcodeSupportedFormats.AZTEC,
        Html5QrcodeSupportedFormats.DATA_MATRIX,
        Html5QrcodeSupportedFormats.PDF_417,
        Html5QrcodeSupportedFormats.MAXICODE,

        // 1D Barcodes
        Html5QrcodeSupportedFormats.CODE_128,
        Html5QrcodeSupportedFormats.CODE_39,
        Html5QrcodeSupportedFormats.CODE_93,
        Html5QrcodeSupportedFormats.CODABAR,

        // EAN/UPC Family (includes ISBN-13)
        Html5QrcodeSupportedFormats.EAN_13,
        Html5QrcodeSupportedFormats.EAN_8,
        Html5QrcodeSupportedFormats.UPC_A,
        Html5QrcodeSupportedFormats.UPC_E,

        // ITF
        Html5QrcodeSupportedFormats.ITF,
        Html5QrcodeSupportedFormats.ITF_14
      ]
    },

    /**
     * Start barcode scanning
     */
    async startScanning () {
      this.scanning = true
      this.scanMessage = 'Kamera wird gestartet...'

      // Check if html5-qrcode is loaded
      if (typeof Html5Qrcode === 'undefined') {
        this.scanMessage = 'Barcode-Library nicht geladen'
        setTimeout(() => { this.scanning = false }, 2000)
        return
      }

      this.html5QrCode = new Html5Qrcode('qr-reader')

      try {
        await this.html5QrCode.start(
          { facingMode: 'environment' }, // Back camera
          {
            fps: 10,
            qrbox: { width: 300, height: 300 },
            formatsToSupport: this.getSupportedFormats(),
            aspectRatio: 1.0,
            disableFlip: false
          },
          (decodedText, decodedResult) => {
            this.onScanSuccess(decodedText, decodedResult)
          },
          (errorMessage) => {
            // Optional: Debug scan attempts
            // console.log('Scan attempt:', errorMessage);
          }
        )

        this.scanMessage = 'Halte den Barcode in den Rahmen'
      } catch (err) {
        console.error('Scanner error:', err)

        let message = 'Kamera-Zugriff fehlgeschlagen'
        if (err.name === 'NotAllowedError') {
          message = 'Kamera-Berechtigung verweigert'
        } else if (err.name === 'NotFoundError') {
          message = 'Keine Kamera gefunden'
        } else if (err.name === 'NotReadableError') {
          message = 'Kamera wird bereits verwendet'
        }

        this.scanMessage = message
        setTimeout(() => { this.scanning = false }, 3000)
      }
    },

    /**
     * Handle successful scan with automatic type detection
     */
    onScanSuccess (decodedText, decodedResult) {
      // Get barcode format name
      const formatName = decodedResult.result.format.formatName

      // Map to internal format name
      const internalType = BARCODE_TYPE_MAPPING[formatName] || formatName

      // Check for ISBN-10 (special case)
      if (this.isISBN10(decodedText)) {
        this[config.fieldName] = this.convertISBN10toISBN13(decodedText)
        this.barcodeType = 'ISBN13'
        this.scanMessage = 'ISBN-10 erkannt (zu ISBN-13 konvertiert)'
      } else {
        this[config.fieldName] = decodedText
        this.barcodeType = internalType
        this.scanMessage = `${formatName} erkannt!`
      }

      // Auto-populate barcode type dropdown
      this.$nextTick(() => {
        this.updateBarcodeTypeDropdown()
      })

      setTimeout(() => this.stopScanning(), 1000)
    },

    /**
     * Update barcode type dropdown if present
     */
    updateBarcodeTypeDropdown () {
      const dropdown = document.querySelector('select[name="barcode_type"]')
      if (dropdown && this.barcodeType) {
        const option = dropdown.querySelector(`option[value="${this.barcodeType}"]`)
        if (option) {
          dropdown.value = this.barcodeType
          // Trigger change event for Alpine.js reactivity
          dropdown.dispatchEvent(new Event('change', { bubbles: true }))
        }
      }
    },

    /**
     * Stop barcode scanning
     */
    stopScanning () {
      if (this.html5QrCode) {
        this.html5QrCode.stop().then(() => {
          this.html5QrCode.clear()
          this.scanning = false
          this.scanMessage = config.defaultMessage
        }).catch(err => {
          console.error('Stop error:', err)
          this.scanning = false
          this.scanMessage = config.defaultMessage
        })
      }
    },

    /**
     * Check if string is ISBN-10 format
     */
    isISBN10 (code) {
      return /^\d{9}[\dX]$/.test(code)
    },

    /**
     * Convert ISBN-10 to ISBN-13 (EAN-13 with 978 prefix)
     */
    convertISBN10toISBN13 (isbn10) {
      // Remove check digit
      const isbn9 = isbn10.substring(0, 9)
      const isbn12 = '978' + isbn9

      // Calculate EAN-13 check digit
      let sum = 0
      for (let i = 0; i < 12; i++) {
        sum += parseInt(isbn12[i]) * (i % 2 === 0 ? 1 : 3)
      }
      const checkDigit = (10 - (sum % 10)) % 10

      return isbn12 + checkDigit
    }
  }
}

/**
 * Card scanner - optimized for common loyalty card formats
 */
window.cardForm = function (initialCardNumber = '') {
  return createBarcodeScanner(
    {
      fieldName: 'cardNumber',
      initialValue: initialCardNumber,
      defaultMessage: 'Halte die Karte in den Rahmen',
      successMessage: 'Barcode erkannt!'
    },
    [
      Html5QrcodeSupportedFormats.CODE_128,
      Html5QrcodeSupportedFormats.EAN_13,
      Html5QrcodeSupportedFormats.EAN_8,
      Html5QrcodeSupportedFormats.QR_CODE,
      Html5QrcodeSupportedFormats.CODE_39,
      Html5QrcodeSupportedFormats.UPC_A
    ]
  )
}

/**
 * Voucher scanner - all formats for maximum flexibility
 */
window.voucherForm = function (initialCode = '') {
  return createBarcodeScanner(
    {
      fieldName: 'voucherCode',
      initialValue: initialCode,
      defaultMessage: 'Halte den Gutschein-Code in den Rahmen',
      successMessage: 'Code erkannt!'
    },
    null // All formats
  )
}

/**
 * Gift Card scanner - common gift card formats
 */
window.giftCardForm = function (initialCardNumber = '') {
  const scanner = createBarcodeScanner(
    {
      fieldName: 'cardNumber',
      initialValue: initialCardNumber,
      defaultMessage: 'Halte die Kartennummer in den Rahmen',
      successMessage: 'Barcode erkannt!'
    },
    [
      Html5QrcodeSupportedFormats.CODE_128,
      Html5QrcodeSupportedFormats.EAN_13,
      Html5QrcodeSupportedFormats.QR_CODE,
      Html5QrcodeSupportedFormats.PDF_417,
      Html5QrcodeSupportedFormats.UPC_A,
      Html5QrcodeSupportedFormats.CODE_39
    ]
  )

  // Add gift card specific field
  scanner.merchantId = ''

  return scanner
}

/**
 * Email Autocomplete for sharing
 */
window.emailAutocomplete = function () {
  return {
    email: '',
    suggestions: [],
    showSuggestions: false,
    loading: false,
    selectedIndex: -1,

    async fetchSuggestions () {
      if (this.email.length < 1) {
        this.suggestions = []
        this.showSuggestions = false
        return
      }

      this.loading = true

      try {
        const response = await fetch(`/api/shared-users?q=${encodeURIComponent(this.email)}`)
        const data = await response.json()
        this.suggestions = data || []
        this.showSuggestions = this.suggestions.length > 0
        this.selectedIndex = -1
      } catch (error) {
        console.error('Failed to fetch suggestions:', error)
        this.suggestions = []
        this.showSuggestions = false
      } finally {
        this.loading = false
      }
    },

    selectSuggestion (suggestion) {
      this.email = suggestion.email
      this.showSuggestions = false
      this.suggestions = []
      this.selectedIndex = -1
    },

    handleKeydown (event) {
      if (!this.showSuggestions || this.suggestions.length === 0) return

      if (event.key === 'ArrowDown') {
        event.preventDefault()
        this.selectedIndex = Math.min(this.selectedIndex + 1, this.suggestions.length - 1)
      } else if (event.key === 'ArrowUp') {
        event.preventDefault()
        this.selectedIndex = Math.max(this.selectedIndex - 1, -1)
      } else if (event.key === 'Enter' && this.selectedIndex >= 0) {
        event.preventDefault()
        this.selectSuggestion(this.suggestions[this.selectedIndex])
      } else if (event.key === 'Escape') {
        this.showSuggestions = false
        this.selectedIndex = -1
      }
    },

    hideSuggestions () {
      // Delay to allow click event to fire first
      setTimeout(() => {
        this.showSuggestions = false
        this.selectedIndex = -1
      }, 200)
    }
  }
}
