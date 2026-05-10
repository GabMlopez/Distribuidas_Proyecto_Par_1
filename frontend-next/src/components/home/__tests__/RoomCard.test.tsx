import { render, screen, fireEvent } from '@testing-library/react'
import { RoomCard } from '../RoomCard'
import { Room } from '@/types'

const mockRoom: Room = {
  sala_id: '12345',
  nombre: 'Sala de Prueba',
  tipo: 'texto',
  pin: '1234',
  usuarios: 2,
}

describe('RoomCard Component', () => {
  const mockOpenJoin = jest.fn()
  const mockOpenEdit = jest.fn()
  const mockDeleteRoom = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders room name correctly', () => {
    render(
      <RoomCard
        room={mockRoom}
        index={0}
        isAdmin={false}
        openJoin={mockOpenJoin}
        openEdit={mockOpenEdit}
        deleteRoom={mockDeleteRoom}
      />
    )

    // Should render the room name twice (one truncated text, one inside tooltip)
    const elements = screen.getAllByText('Sala de Prueba')
    expect(elements.length).toBeGreaterThan(0)
  })

  it('calls openJoin when Unirse is clicked', () => {
    render(
      <RoomCard
        room={mockRoom}
        index={0}
        isAdmin={false}
        openJoin={mockOpenJoin}
        openEdit={mockOpenEdit}
        deleteRoom={mockDeleteRoom}
      />
    )

    const joinButton = screen.getByText('Unirse')
    fireEvent.click(joinButton)

    expect(mockOpenJoin).toHaveBeenCalledTimes(1)
    expect(mockOpenJoin).toHaveBeenCalledWith(mockRoom)
  })

  it('shows PIN when user is admin', () => {
    render(
      <RoomCard
        room={mockRoom}
        index={0}
        isAdmin={true}
        openJoin={mockOpenJoin}
        openEdit={mockOpenEdit}
        deleteRoom={mockDeleteRoom}
      />
    )

    expect(screen.getByText('1234')).toBeInTheDocument()
  })

  it('does not show PIN when user is not admin', () => {
    render(
      <RoomCard
        room={mockRoom}
        index={0}
        isAdmin={false}
        openJoin={mockOpenJoin}
        openEdit={mockOpenEdit}
        deleteRoom={mockDeleteRoom}
      />
    )

    expect(screen.queryByText('1234')).not.toBeInTheDocument()
  })
})
