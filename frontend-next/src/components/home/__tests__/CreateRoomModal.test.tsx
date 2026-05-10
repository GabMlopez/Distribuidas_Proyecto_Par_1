import { render, screen, fireEvent } from '@testing-library/react'
import { CreateRoomModal } from '../CreateRoomModal'

describe('CreateRoomModal Component', () => {
  const mockOnClose = jest.fn()
  const mockSetSalanombre = jest.fn()
  const mockSetNewRoomPin = jest.fn()
  const mockSetNewRoomType = jest.fn()
  const mockCreateRoom = jest.fn()

  const defaultProps = {
    show: true,
    onClose: mockOnClose,
    salanombre: '',
    setSalanombre: mockSetSalanombre,
    newRoomPin: '',
    setNewRoomPin: mockSetNewRoomPin,
    newRoomType: 'texto' as const,
    setNewRoomType: mockSetNewRoomType,
    createRoom: mockCreateRoom,
    creating: false,
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders the modal when show is true', () => {
    render(<CreateRoomModal {...defaultProps} />)
    expect(screen.getByText('Crear nueva sala')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Mínimo 4 dígitos')).toBeInTheDocument()
  })

  it('does not render the modal when show is false', () => {
    render(<CreateRoomModal {...defaultProps} show={false} />)
    expect(screen.queryByText('Crear nueva sala')).not.toBeInTheDocument()
  })

  it('calls onClose when the cancel button is clicked', () => {
    render(<CreateRoomModal {...defaultProps} />)
    const cancelButton = screen.getByText('Cancelar')
    fireEvent.click(cancelButton)
    expect(mockOnClose).toHaveBeenCalledTimes(1)
  })

  it('calls setSalanombre when name input changes', () => {
    render(<CreateRoomModal {...defaultProps} />)
    const nameInput = screen.getByPlaceholderText('Nombre de la sala')
    fireEvent.change(nameInput, { target: { value: 'Sala Test' } })
    expect(mockSetSalanombre).toHaveBeenCalledWith('Sala Test')
  })

  it('calls setNewRoomPin when pin input changes and only passes numbers', () => {
    render(<CreateRoomModal {...defaultProps} />)
    const pinInput = screen.getByPlaceholderText('Mínimo 4 dígitos')
    fireEvent.change(pinInput, { target: { value: '12a34' } })
    expect(mockSetNewRoomPin).toHaveBeenCalledWith('1234')
  })

  it('calls createRoom when submit button is clicked', () => {
    render(<CreateRoomModal {...defaultProps} newRoomPin="1234" />)
    const createButton = screen.getByText('Crear sala')
    fireEvent.click(createButton)
    expect(mockCreateRoom).toHaveBeenCalledTimes(1)
  })

  it('disables submit button if PIN is less than 4 digits', () => {
    render(<CreateRoomModal {...defaultProps} newRoomPin="123" />)
    const createButton = screen.getByText('Crear sala')
    expect(createButton).toBeDisabled()
  })
})
