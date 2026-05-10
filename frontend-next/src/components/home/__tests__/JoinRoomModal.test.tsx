import { render, screen, fireEvent } from '@testing-library/react'
import { JoinRoomModal } from '../JoinRoomModal'

describe('JoinRoomModal Component', () => {
  const mockOnClose = jest.fn()
  const mockSetJoinPin = jest.fn()
  const mockJoinRoom = jest.fn()

  const defaultProps = {
    show: true,
    onClose: mockOnClose,
    selectedRoomName: 'Sala de Prueba 1',
    joinPin: '',
    setJoinPin: mockSetJoinPin,
    joinRoom: mockJoinRoom,
    joining: false,
    joinError: '',
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders correctly when show is true', () => {
    render(<JoinRoomModal {...defaultProps} />)
    expect(screen.getByText('Unirse a la sala')).toBeInTheDocument()
    expect(screen.getByText('Sala de Prueba 1')).toBeInTheDocument()
  })

  it('does not render when show is false', () => {
    render(<JoinRoomModal {...defaultProps} show={false} />)
    expect(screen.queryByText('Unirse a la sala')).not.toBeInTheDocument()
  })

  it('calls setJoinPin when pin input changes', () => {
    render(<JoinRoomModal {...defaultProps} />)
    const pinInput = screen.getByPlaceholderText('••••••')
    fireEvent.change(pinInput, { target: { value: '1234' } })
    expect(mockSetJoinPin).toHaveBeenCalledWith('1234')
  })

  it('calls joinRoom when form is submitted by clicking Entrar', () => {
    render(<JoinRoomModal {...defaultProps} joinPin="1234" />)
    const joinBtn = screen.getByText('Entrar', { selector: 'button' })
    fireEvent.click(joinBtn)
    expect(mockJoinRoom).toHaveBeenCalledTimes(1)
  })

  it('disables the button when pin is less than 4 digits', () => {
    render(<JoinRoomModal {...defaultProps} joinPin="123" />)
    const joinBtn = screen.getByText('Entrar', { selector: 'button' })
    expect(joinBtn).toBeDisabled()
  })

  it('shows an error message if joinError is passed', () => {
    render(<JoinRoomModal {...defaultProps} joinError="PIN incorrecto" />)
    expect(screen.getByText('PIN incorrecto')).toBeInTheDocument()
  })

  it('shows "Conectando..." text when joining is true', () => {
    render(<JoinRoomModal {...defaultProps} joining={true} joinPin="1234" />)
    const joinBtn = screen.getByText('Conectando...', { selector: 'button' })
    expect(joinBtn).toBeInTheDocument()
    expect(joinBtn).toBeDisabled()
  })
})
