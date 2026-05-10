import { render, screen, fireEvent } from '@testing-library/react'
import { DeleteModal } from '../DeleteRoomModal'

describe('DeleteModal Component', () => {
  const mockOnClose = jest.fn()
  const mockOnConfirm = jest.fn()

  const defaultProps = {
    show: true,
    onClose: mockOnClose,
    onConfirm: mockOnConfirm,
    roomName: 'Sala de Prueba',
    roomId: 'room-1',
    deleting: false,
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders correctly when show is true', () => {
    render(<DeleteModal {...defaultProps} />)
    expect(screen.getByText('Eliminar sala')).toBeInTheDocument()
    expect(screen.getByText('Sala de Prueba')).toBeInTheDocument()
    expect(screen.getByText('ID: room-1')).toBeInTheDocument()
  })

  it('does not render when show is false', () => {
    render(<DeleteModal {...defaultProps} show={false} />)
    expect(screen.queryByText('Eliminar sala')).not.toBeInTheDocument()
  })

  it('calls onClose when cancel is clicked', () => {
    render(<DeleteModal {...defaultProps} />)
    fireEvent.click(screen.getByText('Cancelar'))
    expect(mockOnClose).toHaveBeenCalledTimes(1)
  })

  it('calls onConfirm when confirm is clicked', () => {
    render(<DeleteModal {...defaultProps} />)
    
    // El botón tiene el SVG y el texto "Eliminar sala", así que lo buscamos por rol o text content
    const deleteButton = screen.getAllByText('Eliminar sala')[1] // el primero es el h3, el segundo es el botón
    // O mejor buscar por el botón específicamente
    const button = screen.getByRole('button', { name: /Eliminar sala/i })
    fireEvent.click(button)

    expect(mockOnConfirm).toHaveBeenCalledTimes(1)
  })

  it('shows deleting state and disables buttons when deleting is true', () => {
    render(<DeleteModal {...defaultProps} deleting={true} />)
    
    expect(screen.getByText('Eliminando...')).toBeInTheDocument()
    
    const cancelButton = screen.getByText('Cancelar')
    expect(cancelButton).toBeDisabled()
    
    const deleteButton = screen.getByRole('button', { name: /Eliminando.../i })
    expect(deleteButton).toBeDisabled()
  })
})
