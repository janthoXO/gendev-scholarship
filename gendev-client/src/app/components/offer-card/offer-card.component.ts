import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Offer } from '../../models/offer.model';

@Component({
  selector: 'app-offer-card',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './offer-card.component.html',
  styleUrl: './offer-card.component.css'
})
export class OfferCardComponent {
  offer = input.required<Offer>();
}
