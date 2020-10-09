#!/bin/bash
ledger reg -X '$' --register-format '%(format_date(date)),%(quoted(payee)),%(quoted(display_account)),%(quantity(display_amount)),%(quantity(display_total))\n' "$@"
