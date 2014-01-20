/* simulate the Busy Beaver 4 machine */

#include <stdio.h>
#include <stdlib.h>

enum {
    A, B, C, D, H
} states;

enum {
    LEFT, RIGHT
} directions;

typedef struct N_ {
    struct N_* left;
    struct N_* right;
    char value;
} N;

const int mask = 0x3ffffff;

int alloced = 0;
N* leftmost;

#define newN( p, l, r ) \
    p = (N*) malloc( sizeof( N ) ); \
    p->left = l; \
    p->right = r; \
    p->value = 0;

#define move( tape, direction ) \
    if ( direction == LEFT ) { \
        if ( !tape->left ) { \
            newN( tape->left, 0, tape ); \
            leftmost = tape->left; \
        } \
        tape = tape->left; \
    } \
    else { \
        if ( !tape->right ) { \
            newN( tape->right, tape, 0 ); \
            alloced++; \
        } \
        tape = tape->right; \
    }

#define update( S, W, D ) state = S; tape->value = W; move( tape, D )

int
main( int argc, char** argv )
{
    N* tape;
    N* q;
    int i;

    //newN( tape, 0, 0 );
    #define PREALLOC 204000
    tape = (N*) malloc( PREALLOC * sizeof( N ) );
    tape[ 0 ].left = 0;
    tape[ 0 ].right = &tape[ 1 ];
    tape[ 0 ].value = 0;
    for ( i = 1; i < PREALLOC - 1; i++ ) {
        tape[ i ].left = &tape[ i - 1 ];
        tape[ i ].right = &tape[ i + 1 ];
        tape[ i ].value = 0;
    }
    tape[ PREALLOC - 1 ].left = &tape[ PREALLOC - 2 ];
    tape[ PREALLOC - 1 ].right = 0;
    tape[ PREALLOC - 1 ].value = 0;
    alloced = PREALLOC;

    leftmost = tape;
    char state = 0;
    char done = 0;
    unsigned long long steps = 0;
    for ( steps = 0; !done; steps++ ) {
        if ( tape->value ) {
            switch ( state ) {
            case A: update( B, 1, LEFT ); break;
            case B: update( C, 0, LEFT ); break;
            case C: update( D, 1, LEFT ); break;
            case D: update( A, 0, RIGHT ); break;
            }
        }
        else {
            switch ( state ) {
            case A: update( B, 1, RIGHT ); break;
            case B: update( A, 1, LEFT ); break;
            case C: update( H, 1, RIGHT ); done = 1; break;
            case D: update( D, 1, RIGHT ); break;
            }
        }
        if ( ( steps & mask ) == mask )
            printf( "steps=%llu length=%d\n", steps, alloced );
    }

    int ones = 0;
    for ( q = leftmost; q; q = q->right ) {
        if ( q->value )
            ones++;
    }

    printf( "steps=%llu length=%d ones=%d\n", steps, alloced, ones );

    return 0;
}
