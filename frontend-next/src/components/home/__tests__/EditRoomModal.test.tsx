import { render, screen, fireEvent } from '@testing-library/react'
import { EditRoomModal } from '../EditRoomModal'

describe('EditRoomModal Component', () => {
  const mockOnClose = jest.fn()
  const mockSetEditNombre = jest.fn()
  const mockSetEditPin = jest.fn()
  const mockSetEditType = jest.fn()
  const mockUpdateRoom = jest.fn()

  const defaultProps = {
    show: true,
    onClose: mockOnClose,
    selectedRoomId: 'room-1',
    selectedRoomNombre: 'Sala Original',
    editNombre: '',
    setEditNombre: mockSetEditNombre,
    editPin: '',
    setEditPin: mockSetEditPin,
    editType: 'texto' as const,
    setEditType: mockSetEditType,
    updateRoom: mockUpdateRoom,
    updating: false,
    hasChanges: false,
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders correctly when show is true', () => {
    render(<EditRoomModal {...defaultProps} />)
    expect(screen.getByText('Editar sala')).toBeInTheDocument()
    expect(screen.getByText('ID: room-1')).toBeInTheDocument()
  })

  it('disables the save button if there are no changes', () => {
    render(<EditRoomModal {...defaultProps} hasChanges={false} />)
    const saveButton = screen.getByText('Guardar cambios', { selector: 'button' })
    expect(saveButton).toBeDisabled()
  })

  it('enables the save button if there are changes', () => {
    render(<EditRoomModal {...defaultProps} hasChanges={true} />)
    const saveButton = screen.getByText('Guardar cambios', { selector: 'button' })
    expect(saveButton).not.toBeDisabled()
  })

  it('shows an error if PIN is invalid', () => {
    render(<EditRoomModal {...defaultProps} hasChanges={true} editPin="12" />)
    expect(screen.getByText('El PIN debe tener entre 4 y 6 dígitos')).toBeInTheDocument()
    const saveButton = screen.getByText('Guardar cambios', { selector: 'button' })
    expect(saveButton).toBeDisabled()
  })

  it('calls setEditType when type buttons are clicked', () => {
    render(<EditRoomModal {...defaultProps} />)
    const textButton = screen.getByText('Texto', { selector: 'button' })
    const multimediaButton = screen.getByText('Multimedia', { selector: 'button' })
    
    fireEvent.click(textButton)
    expect(mockSetEditType).toHaveBeenCalledWith('texto')

    fireEvent.click(multimediaButton)
    expect(mockSetEditType).toHaveBeenCalledWith('multimedia')
  })

  it('calls updateRoom when form is submitted', () => {
    render(<EditRoomModal {...defaultProps} hasChanges={true} />)
    const saveButton = screen.getByText('Guardar cambios', { selector: 'button' })
    fireEvent.click(saveButton)
    expect(mockUpdateRoom).toHaveBeenCalledTimes(1)
  })
})
