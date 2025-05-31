import { Component, signal, effect, input, output, OnDestroy, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import { HlmButtonDirective } from '@spartan-ng/helm/button';
import {
  faClose,
  faCopy,
  faCheck,
  faExclamationTriangle,
  faSpinner,
} from '@fortawesome/free-solid-svg-icons';
import { SearchService } from '../../services/search.service';
import { FilterOptions } from '../../models/filterOptions.model';
import { State } from '../../services/state';

@Component({
  selector: 'app-share-dialog',
  imports: [CommonModule, FontAwesomeModule, HlmButtonDirective],
  templateUrl: './share-dialog.component.html',
  styleUrl: './share-dialog.component.css',
  standalone: true,
})
export class ShareDialogComponent implements OnDestroy {
  protected faClose = faClose;
  protected faCopy = faCopy;
  protected faCheck = faCheck;
  protected faExclamationTriangle = faExclamationTriangle;
  protected faSpinner = faSpinner;

  show = input<boolean>(false);
  queryHash = input<string | null>(null);
  filters = input<FilterOptions | null>(null);
  close = output<void>();

  shareLink = signal('');
  linkCopied = signal(false);

  error = signal<string | null>(null);
  isGeneratingLink = signal(false);

  constructor(private searchService: SearchService, private state: State) {
    // Watch for show changes to trigger share link generation and manage body scroll
    effect(() => {
      if (this.show() && this.queryHash() && this.state.sessionId()) {
        this.generateShareLink(
          this.queryHash()!,
          this.state.sessionId(),
          this.filters() || undefined
        );
        // Prevent body scrolling when modal is open
        document.body.style.overflow = 'hidden';
      } else if (!this.show()) {
        this.resetState();
        // Restore body scrolling when modal is closed
        document.body.style.overflow = '';
      }
    });
  }

  private generateShareLink(
    queryHash: string,
    sessionId: string,
    filters?: FilterOptions
  ) {
    this.isGeneratingLink.set(true);
    this.error.set(null);
    this.linkCopied.set(false);

    this.searchService.shareOffers(queryHash, sessionId, filters).subscribe({
      next: (shareId: string) => {
        const shareUrl = `${window.location.origin}/offers/shared/${shareId}`;
        this.shareLink.set(shareUrl);
      },
      error: (error) => {
        console.error('Failed to create share link:', error);
        this.error.set('Failed to create share link. Please try again.');
        this.isGeneratingLink.set(false);
      },
      complete: () => {
        this.isGeneratingLink.set(false);
      },
    });
  }

  copyShareLink() {
    const link = this.shareLink();
    if (link) {
      navigator.clipboard
        .writeText(link)
        .then(() => {
          this.linkCopied.set(true);
        })
        .catch((err) => {
          console.error('Failed to copy link:', err);
          this.error.set('Failed to copy link to clipboard.');
        });
    }
  }

  closeDialog() {
    this.close.emit();
  }

  retryGeneration() {
    if (this.queryHash() && this.state.sessionId()) {
      this.generateShareLink(
        this.queryHash()!,
        this.state.sessionId(),
        this.filters() || undefined
      );
    }
  }

  private resetState() {
    this.shareLink.set('');
    this.linkCopied.set(false);
    this.error.set(null);
    this.isGeneratingLink.set(false);
  }

  ngOnDestroy() {
    // Ensure body scrolling is restored when component is destroyed
    document.body.style.overflow = '';
  }

  @HostListener('document:keydown.escape', ['$event'])
  onEscapeKey(event: KeyboardEvent) {
    if (this.show()) {
      this.closeDialog();
    }
  }
}
