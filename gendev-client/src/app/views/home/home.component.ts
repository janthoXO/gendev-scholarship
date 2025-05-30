import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { SearchInputComponent } from './search-input/search-input.component';
import { State } from '../../services/state';
import { SearchService } from '../../services/search.service';
import { Address } from '../../models/address.model';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, SearchInputComponent],
  templateUrl: './home.component.html',
  styleUrl: './home.component.css'
})
export class HomeComponent {
  constructor(
    private router: Router,
    private searchStateService: State,
    private searchService: SearchService
  ) {}

  onSearchSubmitted(address: Address) {
    // Navigate to offers page with address as query parameters
    this.router.navigate(['/offers'], {
      queryParams: {
        street: address.street,
        houseNumber: address.houseNumber,
        city: address.city,
        zipCode: address.zipCode
      }
    });
  }
}
