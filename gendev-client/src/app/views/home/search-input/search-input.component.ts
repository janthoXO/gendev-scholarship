import { Component, signal, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Address } from '../../../models/address.model';

@Component({
  selector: 'app-search-input',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './search-input.component.html',
  styleUrl: './search-input.component.css'
})
export class SearchInputComponent {
  // Signals for form state
  isExpanded = signal(false);
  street = signal('');
  houseNumber = signal('');
  city = signal('');
  zipCode = signal('');

  // Output event
  searchSubmitted = output<Address>();

  toggleExpanded() {
    this.isExpanded.update(expanded => !expanded);
  }

  isFormValid(): boolean {
    return this.street().trim() !== '' &&
           this.houseNumber().trim() !== '' &&
           this.city().trim() !== '' &&
           this.zipCode().trim() !== '';
  }

  onSubmit() {
    if (this.isFormValid()) {
      const address: Address = {
        street: this.street().trim(),
        houseNumber: this.houseNumber().trim(),
        city: this.city().trim(),
        zipCode: this.zipCode().trim()
      };
      
      this.searchSubmitted.emit(address);
    }
  }
}
