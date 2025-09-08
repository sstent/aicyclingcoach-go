import { render, screen, fireEvent } from '@testing-library/react'
import FileUpload from '../FileUpload'

describe('FileUpload Component', () => {
  test('renders upload button', () => {
    render(<FileUpload onUpload={() => {}} />)
    expect(screen.getByText('Upload GPX File')).toBeInTheDocument()
    expect(screen.getByTestId('file-input')).toBeInTheDocument()
  })

  test('handles file selection', () => {
    const mockFile = new File(['test content'], 'test.gpx', { type: 'application/gpx+xml' })
    const mockOnUpload = jest.fn()
    
    render(<FileUpload onUpload={mockOnUpload} />)
    
    const input = screen.getByTestId('file-input')
    fireEvent.change(input, { target: { files: [mockFile] } })
    
    expect(mockOnUpload).toHaveBeenCalledWith(mockFile)
    expect(screen.getByText('Selected file: test.gpx')).toBeInTheDocument()
  })

  test('shows error for invalid file type', () => {
    const invalidFile = new File(['test'], 'test.txt', { type: 'text/plain' })
    const { container } = render(<FileUpload onUpload={() => {}} />)
    
    const input = screen.getByTestId('file-input')
    fireEvent.change(input, { target: { files: [invalidFile] } })
    
    expect(screen.getByText('Invalid file type. Please upload a GPX file.')).toBeInTheDocument()
    expect(container.querySelector('.error-message')).toBeVisible()
  })
})